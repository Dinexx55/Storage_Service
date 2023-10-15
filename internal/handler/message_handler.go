package handler

import (
	"StorageService/internal/model"
	"StorageService/internal/service"
	"encoding/json"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type StoreService interface {
	CreateStore(data service.Store, login string) error
	CreateStoreVersion(data service.StoreVersion, storeId, login string) error
	DeleteStore(storeId, login string) error
	DeleteStoreVersion(storeId, versionId, login string) error
	GetStoreByID(storeId, login string) (*model.Store, error)
	GetStoreVersionHistory(storeId, login string) ([]*model.StoreVersion, error)
	GetStoreVersionByID(storeId, versionId, login string) (*model.StoreVersion, error)
}

type StoreFromMessage struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	OwnerName   string `json:"ownerName" binding:"required"`
	OpeningTime string `json:"openingTime" binding:"required"`
	ClosingTime string `json:"closingTime" binding:"required"`
}

type StoreVersionFromMessage struct {
	StoreOwnerName string `json:"storeOwnerName" binding:"required"`
	OpeningTime    string `json:"openingTime" binding:"required"`
	ClosingTime    string `json:"closingTime" binding:"required"`
}

type Message struct {
	Action    string          `json:"action"`
	Data      json.RawMessage `json:"data"`
	StoreID   string          `json:"storeId"`
	UserLogin string          `json:"userLogin"`
	VersionID string          `json:"versionId"`
}

type MessageHandler struct {
	storeService StoreService
	logger       *zap.Logger
}

func NewMessageHandler(storeService StoreService, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		storeService: storeService,
		logger:       logger,
	}
}

func (h *MessageHandler) HandleMessage(msg amqp.Delivery) {
	h.logger.Info("Received message", zap.ByteString("message", msg.Body))

	userLogin := extractLogin(msg)
	action := extractAction(msg)

	switch action {
	case "delete_store":
		storeId := extractStoreID(msg)
		err := h.storeService.DeleteStore(storeId, userLogin)
		if err != nil {
			h.logger.Error("Failed to delete store", zap.Error(err))
		} else {
			h.logger.Info("StoreFromMessage deleted successfully")
		}
		break
	case "delete_store_version":
		storeId := extractStoreID(msg)
		versionId := extractVersionID(msg)
		err := h.storeService.DeleteStoreVersion(storeId, versionId, userLogin)
		if err != nil {
			h.logger.Error("Failed to delete store version", zap.Error(err))

		} else {
			h.logger.Info("StoreFromMessage version deleted successfully")
		}
		break
	case "create_store":
		storeData, err := extractStoreData(msg)

		if err != nil {
			h.logger.Error("Failed to extract data", zap.Error(err))
			return
		}

		srvStore := service.Store{
			Name:        storeData.Name,
			Address:     storeData.Address,
			OwnerName:   storeData.OwnerName,
			OpeningTime: storeData.OpeningTime,
			ClosingTime: storeData.ClosingTime,
		}

		err = h.storeService.CreateStore(srvStore, userLogin)
		if err != nil {
			h.logger.Error("Failed to create store", zap.Error(err))

		} else {
			h.logger.Info("StoreFromMessage created successfully")
		}
		break
	case "create_store_version":
		storeId := extractStoreID(msg)
		storeVersionData, err := extractStoreVersionData(msg)

		if err != nil {
			h.logger.Error("Failed to extract data", zap.Error(err))
			return
		}

		srvStoreVersion := service.StoreVersion{
			StoreOwnerName: storeVersionData.StoreOwnerName,
			OpeningTime:    storeVersionData.OpeningTime,
			ClosingTime:    storeVersionData.ClosingTime,
		}

		err = h.storeService.CreateStoreVersion(srvStoreVersion, storeId, userLogin)
		if err != nil {
			h.logger.Error("Failed to create store", zap.Error(err))

		} else {
			h.logger.Info("StoreFromMessage created successfully")
		}
		break
	case "get_store":
		storeId := extractStoreID(msg)
		store, err := h.storeService.GetStoreByID(storeId, userLogin)
		if err != nil {
			h.logger.Error("Failed to get store", zap.Error(err))

		} else {
			h.logger.Info("Successfully got the store", zap.Any("store", store))
		}
		break
	case "get_store_history":
		storeId := extractStoreID(msg)
		storeHistory, err := h.storeService.GetStoreVersionHistory(storeId, userLogin)
		if err != nil {
			h.logger.Error("Failed to create store", zap.Error(err))

		} else {
			h.logger.Info("Successfully got the store", zap.Any("store", storeHistory))
		}
		break
	case "get_store_version":
		storeId := extractStoreID(msg)
		versionId := extractVersionID(msg)
		storeVersion, err := h.storeService.GetStoreVersionByID(storeId, versionId, userLogin)
		if err != nil {
			h.logger.Error("Failed to create store", zap.Error(err))

		} else {
			h.logger.Info("Successfully got the store", zap.Any("store", storeVersion))
		}
		break
	default:
		h.logger.Warn("Unknown action", zap.String("action", action))
	}
}

func extractStoreID(msg amqp.Delivery) string {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return ""
	}
	return message.StoreID
}

func extractVersionID(msg amqp.Delivery) string {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		// Обработка ошибки
		return ""
	}
	return message.VersionID
}

func extractAction(msg amqp.Delivery) string {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return ""
	}
	return message.Action
}

func extractStoreData(msg amqp.Delivery) (StoreFromMessage, error) {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return StoreFromMessage{}, err
	}

	var storeData StoreFromMessage
	err = json.Unmarshal(message.Data, &storeData)
	if err != nil {
		return StoreFromMessage{}, err
	}

	return storeData, nil
}

func extractStoreVersionData(msg amqp.Delivery) (StoreVersionFromMessage, error) {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return StoreVersionFromMessage{}, err
	}

	var storeVersionData StoreVersionFromMessage
	err = json.Unmarshal(message.Data, &storeVersionData)
	if err != nil {
		return StoreVersionFromMessage{}, err
	}

	return storeVersionData, nil
}

func extractLogin(msg amqp.Delivery) string {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return ""
	}
	return message.UserLogin
}
