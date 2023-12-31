package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/Lxxx-go/Dou-Yin/app/user/dao"
	"github.com/Lxxx-go/Dou-Yin/models"
	"github.com/Lxxx-go/Dou-Yin/repo"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// [增操作]用户 关注别人，给定UserID和ToUserID，增加一行数据
func FollowUser(ctx context.Context, userID, toUserID int64) error {
	if userID == toUserID {
		return errors.New("不能关注自己")
	}

	// 检查是否已经关注过用户
	var tmp repo.Relation
	err := repo.DB.WithContext(ctx).Table("relation").Where("user_id = ? AND to_user_id = ?", userID, toUserID).First(&tmp).Error
	if err == nil {
		return errors.New("已经关注过该用户")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	relation := repo.Relation{
		UserId:   userID,
		ToUserId: toUserID,
	}

	// 开启事务：
	err = repo.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建关注关系
		if err := tx.Table("relation").Create(&relation).Error; err != nil {
			zap.L().Error("关注别人，MySQL数据库添加关注关系失败: ", zap.Error(err))
			return err
		}

		// 更新关注者的关注数
		if err := tx.Table("user").Where("user_id = ?", userID).Update("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
			zap.L().Error("关注别人，MySQL数据库更新关注者的关注数失败: ", zap.Error(err))
			return err
		}

		// 更新被关注者的粉丝数
		if err := tx.Table("user").Where("user_id = ?", toUserID).Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
			zap.L().Error("关注别人，MySQL数据库更新被关注者的粉丝数失败: ", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		zap.L().Error("MySQL关注别人失败: ", zap.Error(err))
		return err
	}

	zap.L().Info("MySQL成功关注别人")
	return nil
}

// [删操作]用户 取关别人，给定UserID和ToUserID，软删除一行数据
func UnFollowUser(ctx context.Context, userID, toUserID int64) error {
	if userID == toUserID {
		return errors.New("不能取关自己")
	}

	// 检查是否关注过用户 前提：数据库中必须有该条数据
	var temp repo.Relation
	err := repo.DB.WithContext(ctx).Table("relation").Where("user_id = ? AND to_user_id = ?", userID, toUserID).First(&temp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("MySQL数据库中查询到 用户没有关注过该用户，不能取关: ", zap.Error(err))
		return errors.New("用户没有关注过该用户，不能取关")
	} else if err != nil {
		zap.L().Error("MySQL数据库查询 关注关系失败 ", zap.Error(err))
		return err
	}

	relation := repo.Relation{
		UserId:   userID,
		ToUserId: toUserID,
	}

	// 开启事务：
	err = repo.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新关注者的关注数
		if err := tx.Table("user").Where("user_id = ?", userID).Update("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
			zap.L().Error("取关别人，MySQL数据库更新关注者的关注数失败: ", zap.Error(err))
			return err
		}

		// 更新被关注者的粉丝数
		if err := tx.Table("user").Where("user_id = ?", toUserID).Update("follower_count", gorm.Expr("follower_count - ?", 1)).Error; err != nil {
			zap.L().Error("取关别人，MySQL数据库更新被关注者的粉丝数失败: ", zap.Error(err))
			return err
		}

		// 删除关注关系
		if err := tx.Table("relation").Where("user_id = ? AND to_user_id = ?", userID, toUserID).Delete(&relation).Error; err != nil {
			zap.L().Error("取关别人，MySQL数据库软删除关注关系失败: ", zap.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		zap.L().Error("MySQL数据库取关别人失败: ", zap.Error(err))
		return err
	}

	zap.L().Info("MySQL数据库成功取关别人")

	return nil
}

// [查询操作]用户 获取关注人列表: 给定UserID，查询表中所有的userid字段为当前UserID的数据，返回用户的关注者ID切片ToUserIDs[]
func GetFollowList(ctx context.Context, userID int64) ([]int64, error) {
	zap.L().Info("数据库开始获取关注人列表")
	var followList []repo.Relation
	// 在 relation 表中查找所有 user_id 字段等于给定 userID 的数据
	err := repo.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&followList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("MySQL数据库中查询到该用户没有关注者: ", zap.Error(err))
		return make([]int64, 0), err
	}
	if err != nil {
		zap.L().Error("MySQL数据库中查询该用户的关注者失败: ", zap.Error(err))
		return nil, err
	}
	// 提取UserID的所有关注关系中的 关注ID
	toUserIDs := make([]int64, len(followList))
	for i, relation := range followList {
		toUserIDs[i] = int64(relation.ToUserId)
	}

	zap.L().Info("MySQL数据库成功获取关注列表")
	return toUserIDs, nil
}

// [查询操作]用户 获取粉丝列表: 给定UserID，查询表中所有的to_user_id字段为当前UserID的数据，返回用户的粉丝ID切片UserIDs[]
func GetFollowerList(ctx context.Context, userID int64) ([]int64, error) {
	zap.L().Info("数据库开始获取粉丝列表")
	var followerList []repo.Relation

	// 在 relation 表中查找所有 to_user_id 字段等于给定 userID 的数据
	result := repo.DB.WithContext(ctx).Where("to_user_id = ?", userID).Find(&followerList)
	if result.RowsAffected == 0 {
		zap.L().Error("MySQL数据库中查询到该用户没有粉丝")
		return make([]int64, 0), nil // 返回一个空切片，表示没有粉丝
	}
	if result.Error != nil {
		zap.L().Error("MySQL数据库中查找该用户的粉丝失败: ", zap.Error(result.Error))
		return nil, result.Error
	}

	// 提取UserID的所有被关注关系中的 粉丝ID
	userIDs := make([]int64, len(followerList))
	for i, relation := range followerList {
		userIDs[i] = int64(relation.UserId)
	}

	fmt.Printf("获取到的userIDs: %v\n", userIDs)
	zap.L().Info("MySQL数据库成功获取粉丝列表")
	return userIDs, nil
}

// [查询操作]获取好友列表(互关): 给定UserID，查询表中所有和当前UserID互关的用户ID切片
func GetFriendList(ctx context.Context, userID int64) ([]int64, error) {
	zap.L().Info("数据库开始获取好友列表")
	fmt.Println("数据库开始获取好友列表")

	var friendList []repo.Relation

	// 在 relation 表中查找当前UserID互关的数据
	err := repo.DB.WithContext(ctx).Where("user_id = ? AND to_user_id IN (?)", userID, repo.DB.WithContext(ctx).Table("relation").Select("user_id").Where("to_user_id = ?", userID)).Find(&friendList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("MySQL数据库中查询到该用户没有好友: ", zap.Error(err))
		return make([]int64, 0), err
	}
	if err != nil {
		zap.L().Error("MySQL数据库中查找该用户的好友失败: ", zap.Error(err))
		return nil, err
	}
	userIDS := make([]int64, len(friendList))
	for i, fan := range friendList {
		userIDS[i] = int64(fan.UserId)
	}

	zap.L().Info("MySQL数据库成功获取好友列表")
	return userIDS, nil
}

func GetUsersByIDList(ctx context.Context, UserID int64, userIDs []int64) ([]*models.User, error) {
	var users []*models.User
	for _, toUserID := range userIDs {
		userInfo, err := dao.GetuserInfoByID(ctx, toUserID)
		userInfo.IsFollow, _ = dao.IsFollowByID(ctx, UserID, toUserID)
		users = append(users, userInfo)

		if err != nil {
			return nil, err
		}
		fmt.Printf("user_id: %v\n", UserID)
		fmt.Printf("to_user_id: %v\n", toUserID)
		fmt.Printf("isfollow: %v\n", userInfo.IsFollow)
	}
	return users, nil
}

// IsFollowByID 判断是否关注了该用户
func IsFollowByID(ctx context.Context, userID, autherID int64) (bool, error) {
	var rel repo.Relation
	result := repo.DB.WithContext(ctx).Table("relation").Where("user_id = ? AND to_user_id = ?", userID, autherID).Limit(1).Find(&rel)
	if result.Error != nil {
		zap.L().Info("查找关注关系时出错")
		return false, result.Error
	}
	if result.RowsAffected > 0 { //关注了该用户
		return true, nil
	}
	return false, nil //未关注
}

func GetFriendsByIDList(ctx context.Context, UserID int64, friendIDs []int64) ([]*models.FriendUser, error) {
	var friendInfoList []*models.FriendUser
	for _, friendID := range friendIDs {
		// 查询当前用户发送的最新消息
		var sentMessage repo.Message
		err := repo.DB.Order("create_time desc").Where("from_user_id = ? AND to_user_id = ?", UserID, friendID).First(&sentMessage).Error
		if err != nil {
			zap.L().Error("查不到用户发送的最新消息:", zap.Error(err))
		}
		// 查询当前用户接收的最新消息
		var receivedMessage repo.Message
		err = repo.DB.Order("create_time desc").Where("from_user_id = ? AND to_user_id = ?", friendID, UserID).First(&receivedMessage).Error
		if err != nil {
			zap.L().Error("查不到用户接收的最新消息:", zap.Error(err))
		}

		// 获取每个(userId, friendId) 的最新消息 和 消息类型，判断哪个消息更新
		var latestMessage repo.Message
		var isSender int64
		if sentMessage.CreateTime > receivedMessage.CreateTime {
			latestMessage = sentMessage
			isSender = 1
		} else {
			latestMessage = receivedMessage
			isSender = 0
		}
		msgInfo := &models.FriendUser{
			Msg:     latestMessage.Contents, // 最新聊天消息
			MsgType: isSender,               // 消息类型
		}
		friendInfoList = append(friendInfoList, msgInfo)
	}
	return friendInfoList, nil
}
