package handler

import (
	"StorageService/internal/model"
	"StorageService/internal/service"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"net/http"
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
	OwnerName   string `json:"ownerName" binding:"required"`
	OpeningTime string `json:"openingTime" binding:"required"`
	ClosingTime string `json:"closingTime" binding:"required"`
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
	gatewayUrl   string
	logger       *zap.Logger
}

func NewMessageHandler(storeService StoreService, gatewayUrl string, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		storeService: storeService,
		gatewayUrl:   gatewayUrl,
		logger:       logger,
	}
}

func sendResponseToGateway(url string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (h *MessageHandler) HandleMessage(msg amqp.Delivery) {
	h.logger.Info("Received message", zap.ByteString("message", msg.Body))

	userLogin := extractLogin(msg)
	action := extractAction(msg)

	switch action {
	case "delete_store":
		h.handleDeleteStore(msg, userLogin)
	case "delete_store_version":
		h.handleDeleteStoreVersion(msg, userLogin)
	case "create_store":
		h.handleCreateStore(msg, userLogin)
	case "create_store_version":
		h.handleCreateStoreVersion(msg, userLogin)
	case "get_store":
		h.handleGetStore(msg, userLogin)
	case "get_store_history":
		h.handleGetStoreHistory(msg, userLogin)
	case "get_store_version":
		h.handleGetStoreVersion(msg, userLogin)
	default:
		h.logger.Warn("Unknown action", zap.String("action", action))
	}
}

func (h *MessageHandler) handleDeleteStore(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)

	err := h.storeService.DeleteStore(storeId, userLogin)

	if err != nil {
		h.logger.Error("Failed to delete store", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Store deleted successfully")

		err = sendSuccessResponseToGateway(h.gatewayUrl, "Store deleted successfully")
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleDeleteStoreVersion(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)
	versionId := extractVersionID(msg)

	err := h.storeService.DeleteStoreVersion(storeId, versionId, userLogin)

	if err != nil {
		h.logger.Error("Failed to delete store version", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Store version deleted successfully")

		err = sendSuccessResponseToGateway(h.gatewayUrl, "Store version deleted successfully")
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleCreateStore(msg amqp.Delivery, userLogin string) {
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

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Store created successfully")

		err = sendSuccessResponseToGateway(h.gatewayUrl, "Store created successfully")
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleCreateStoreVersion(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)
	storeVersionData, err := extractStoreVersionData(msg)
	if err != nil {
		h.logger.Error("Failed to extract data", zap.Error(err))
		return
	}

	srvStoreVersion := service.StoreVersion{
		OwnerName:   storeVersionData.OwnerName,
		OpeningTime: storeVersionData.OpeningTime,
		ClosingTime: storeVersionData.ClosingTime,
	}

	err = h.storeService.CreateStoreVersion(srvStoreVersion, storeId, userLogin)
	if err != nil {
		h.logger.Error("Failed to create store version", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Store version created successfully")

		err = sendSuccessResponseToGateway(h.gatewayUrl, "Store version created successfully")
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleGetStore(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)
	store, err := h.storeService.GetStoreByID(storeId, userLogin)
	if err != nil {
		h.logger.Error("Failed to get store", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Successfully got the store", zap.Any("store", store))

		err = sendSuccessResponseToGateway(h.gatewayUrl, store)
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleGetStoreHistory(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)
	storeHistory, err := h.storeService.GetStoreVersionHistory(storeId, userLogin)
	if err != nil {
		h.logger.Error("Failed to get store history", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Successfully got the version history", zap.Any("store", storeHistory))

		err = sendSuccessResponseToGateway(h.gatewayUrl, storeHistory)
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
	}
}

func (h *MessageHandler) handleGetStoreVersion(msg amqp.Delivery, userLogin string) {
	storeId := extractStoreID(msg)
	versionId := extractVersionID(msg)
	storeVersion, err := h.storeService.GetStoreVersionByID(storeId, versionId, userLogin)
	if err != nil {
		h.logger.Error("Failed to get store version", zap.Error(err))

		err = sendErrorResponseToGateway(h.gatewayUrl, err.Error())
		if err != nil {
			h.logger.Error("Failed to send error response to Gateway Service", zap.Error(err))
		}
	} else {
		h.logger.Info("Successfully got the store version", zap.Any("store", storeVersion))

		err = sendSuccessResponseToGateway(h.gatewayUrl, storeVersion)
		if err != nil {
			h.logger.Error("Failed to send success response to Gateway Service", zap.Error(err))
		}
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

func sendErrorResponseToGateway(url string, errorMessage interface{}) error {
	errorPayload := map[string]interface{}{
		"error": errorMessage,
	}
	return sendResponseToGateway(url, errorPayload)
}

func sendSuccessResponseToGateway(url string, successMessage interface{}) error {
	successPayload := map[string]interface{}{
		"message": successMessage,
	}
	return sendResponseToGateway(url, successPayload)
}
