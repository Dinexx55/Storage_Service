package service

import (
	"StorageService/internal/model"
	"go.uber.org/zap"
	"time"
)

type Repository interface {
	CreateStore(store model.Store) error
	CreateStoreVersion(storeVersion model.StoreVersion) error
	DeleteStore(storeId string) error
	DeleteStoreVersion(versionId string) error
	GetStoreByID(storeId string) (*model.Store, error)
	GetStoreVersionHistory(storeId string) ([]*model.StoreVersion, error)
	GetStoreVersionByID(versionId string) (*model.StoreVersion, error)
	GetStoreVersionForStore(storeId, versionId string) (*model.StoreVersion, error)
}

type Store struct {
	Name        string
	Address     string
	OwnerName   string
	OpeningTime string
	ClosingTime string
}

type StoreVersion struct {
	OwnerName   string
	OpeningTime string
	ClosingTime string
	CreatedAt   string
}

type StoreService struct {
	logger     *zap.Logger
	repository Repository
}

func NewStoreService(logger *zap.Logger, repository Repository) *StoreService {
	return &StoreService{
		logger:     logger,
		repository: repository,
	}
}

func (s *StoreService) CreateStore(data Store, login string) error {
	storeModel := model.Store{
		Name:         data.Name,
		Address:      data.Address,
		CreatorLogin: login,
		OwnerName:    data.OwnerName,
		OpeningTime:  data.OpeningTime,
		ClosingTime:  data.ClosingTime,
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	err := s.repository.CreateStore(storeModel)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to create store")
		return err
	}
	return nil
}

func (s *StoreService) CreateStoreVersion(data StoreVersion, storeID, login string) error {

	storeVersionModel := model.StoreVersion{
		StoreID:       storeID,
		VersionNumber: 0,
		CreatorLogin:  login,
		OwnerName:     data.OwnerName,
		OpeningTime:   data.OpeningTime,
		ClosingTime:   data.ClosingTime,
		CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
		IsLast:        true,
	}
	err := s.repository.CreateStoreVersion(storeVersionModel)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to create store version")
		return err
	}
	return nil

}

func (s *StoreService) DeleteStore(storeID, login string) error {

	_, err := s.repository.GetStoreByID(storeID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to get store")
		return err
	}

	err = s.repository.DeleteStore(storeID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to delete store")
		return err
	}
	return nil
}

func (s *StoreService) DeleteStoreVersion(storeID, versionID, login string) error {

	_, err := s.repository.GetStoreVersionForStore(storeID, versionID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to get store")
		return err
	}

	err = s.repository.DeleteStoreVersion(versionID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to delete store version")
		return err
	}
	return nil
}

func (s *StoreService) GetStoreByID(storeID, login string) (*model.Store, error) {
	store, err := s.repository.GetStoreByID(storeID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to get store")
		return nil, err
	}
	return store, nil
}

func (s *StoreService) GetStoreVersionHistory(storeID, login string) ([]*model.StoreVersion, error) {
	store, err := s.repository.GetStoreVersionHistory(storeID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to get store version")
		return nil, err
	}
	return store, nil

}

func (s *StoreService) GetStoreVersionByID(storeID, versionID, login string) (*model.StoreVersion, error) {
	store, err := s.repository.GetStoreVersionByID(versionID)
	if err != nil {
		s.logger.With(
			zap.String("place", "service"),
			zap.Error(err),
		).Error("Failed to create store version")
		return nil, err
	}
	return store, nil
}
