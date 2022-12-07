package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
)

type SearchService interface {
	GetCodeByNumber(ctx context.Context, req string) (string, error)
}

func (svc *searchService) SetCodesByPrefix(prefix string, codes []string) {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()
	svc.data[prefix] = codes
}

func (svc *searchService) GetCodesByPrefix(prefix string) ([]string, error) {
	svc.mtx.RLock()
	defer svc.mtx.RUnlock()
	v, ok := svc.data[prefix]
	if !ok {
		return nil, fmt.Errorf("value not found by key: %v", prefix)
	}
	return v, nil
}

func (svc *searchService) GetCodeByNumber(_ context.Context, number string) (string, error) {
	var res string

	for i := 0; i <= len(number)-1; i++ {
		prefix := number[:len(number)-i]
		codes, err := svc.GetCodesByPrefix(prefix)
		if err != nil {
			continue
		}
		if len(codes) == 0 {
			continue
		}
		res = codes[0]
		break
	}

	if len(res) == 0 {
		return res, fmt.Errorf("value not found for number: %v", number)
	}
	return res, nil
}

func (svc *searchService) initialFilling() error {
	filePath := "./data/test_data.csv"
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Unable to read input file %s: %v", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("Unable to parse file as CSV for %s: %v", filePath, err)
	}

	for _, line := range records {
		prefix := line[0]
		codes := strings.ReplaceAll(line[1], " ", "")
		svc.SetCodesByPrefix(prefix, strings.Split(codes, ","))
	}
	return nil
}

type searchService struct {
	mtx  sync.RWMutex
	data map[string][]string
}

func NewSearchService() (*searchService, error) {
	svc := searchService{data: make(map[string][]string)}
	err := svc.initialFilling()
	if err != nil {
		return nil, fmt.Errorf("Failed to initial filling: %v", err)
	}
	return &svc, nil
}
