package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

type FileStorage struct {
	*MemStorage
	FilePath      string
	StoreInterval time.Duration
	Restore       bool
	SyncWrite     bool
	StopChan      chan struct{}
	MU            sync.RWMutex
}

func NewFileStorage(filePath string, storeInterval time.Duration, restore bool) *FileStorage {
	memStorage := NewMemStorage()
	filestorage := &FileStorage{
		MemStorage:    memStorage,
		FilePath:      filePath,
		StoreInterval: storeInterval,
		SyncWrite:     storeInterval == 0,
		StopChan:      make(chan struct{}),
	}

	if restore {
		if err := filestorage.LoadFromFile(); err != nil {
			log.Printf("error loading the data from file: %s: %v", filePath, err)
		} else {
			log.Printf("succesfully loaded data from file: %s", filePath)
		}

	}

	return filestorage

}

func (fs *FileStorage) LoadFromFile() error {
	data, err := os.ReadFile(fs.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("file with the metrics doesnt exists: %s: %v", fs.FilePath, err)
			return nil
		}
		return fmt.Errorf("problem opening file: %w", err)
	}
	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("problem unmarshalling metrics: %w", err)
	}
	fs.MU.Lock()
	defer fs.MU.Unlock()

	fs.counters = make(map[string]int64)
	fs.gauges = make(map[string]float64)

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				fs.gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				fs.counters[metric.ID] = *metric.Delta
			}

		}

	}
	log.Printf("loaded %d metrics from the file: %s", len(metrics), fs.FilePath)
	return nil
}

func (fs *FileStorage) SaveToFile() error {
	fs.MU.Lock()
	metrics := fs.MemStorage.GetAllMetrics()
	fs.MU.Unlock()

	data, err := json.MarshalIndent(metrics, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(fs.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}
	tempfile := fs.FilePath + ".tmp"
	if err := os.WriteFile(tempfile, data, 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	if err := os.Rename(tempfile, fs.FilePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	log.Printf("metrics saved to file: %s", fs.FilePath)
	return nil
}

func (fs *FileStorage) UpdateGauge(name string, value float64) error {
	fs.MU.Lock()
	err := fs.MemStorage.UpdateGauge(name, value)
	fs.MU.Unlock()
	if err != nil {
		return err
	}

	if fs.SyncWrite {
		return fs.SaveToFile()
	}
	return nil
}

func (fs *FileStorage) UpdateCounter(name string, value int64) error {
	fs.MU.Lock()
	err := fs.MemStorage.UpdateCounter(name, value)
	fs.MU.Unlock()

	if err != nil {
		return err
	}

	if fs.SyncWrite {
		return fs.SaveToFile()
	}
	return nil
}

func (fs *FileStorage) GetCounter(name string) (int64, bool) {
	fs.MU.RLock()
	defer fs.MU.RUnlock()
	return fs.MemStorage.GetCounter(name)
}

func (fs *FileStorage) GetGauge(name string) (float64, bool) {
	fs.MU.RLock()
	defer fs.MU.RUnlock()
	return fs.MemStorage.GetGauge(name)
}

func (fs *FileStorage) GetMetric(name string) (*models.Metrics, bool) {
	fs.MU.RLock()
	defer fs.MU.RUnlock()
	return fs.MemStorage.GetMetric(name)
}

func (fs *FileStorage) GetAllMetrics() []models.Metrics {
	fs.MU.RLock()
	defer fs.MU.RUnlock()
	return fs.MemStorage.GetAllMetrics()
}

func (fs *FileStorage) AutoSave() {
	go func() {
		ticker := time.NewTicker(fs.StoreInterval)
		for {
			select {
			case <-ticker.C:
				if err := fs.SaveToFile(); err != nil {
					log.Printf("problem while saving to file:%s", err)
				}
			case <-fs.StopChan:
				if err := fs.SaveToFile(); err != nil {
					log.Printf("problem while saving to file:%s", err)
				}
				return
			}
		}

	}()
}

func (fs *FileStorage) Close() error {
	if !fs.SyncWrite {
		close(fs.StopChan)
	}
	return fs.SaveToFile()
}
