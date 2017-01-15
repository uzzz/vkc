package vkc

import (
	"fmt"
	"strconv"
)

const (
	maxVideoPerRequest = 200
)

type Video struct {
	Title string `json:"title"`
}

type videoService struct {
	client *VkClient
}

type videoGetResponse struct {
	Count int
	Items []*Video
}

func (s *videoService) Get(userId, offset, count int) ([]*Video, error) {
	if count > maxVideoPerRequest {
		return nil, fmt.Errorf("count should be <= %d", maxVideoPerRequest)
	}

	params := map[string]string{
		"owner_id": strconv.Itoa(userId),
		"offset":   strconv.Itoa(offset),
		"count":    strconv.Itoa(count),
	}
	req, err := s.client.NewRequest("video.get", params)
	if err != nil {
		return nil, err
	}

	var response videoGetResponse
	_, err = s.client.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (s *videoService) GetAll(userId int) ([]*Video, error) {
	offset := 0
	videos := make([]*Video, 0)

	for {
		videosBatch, err := s.Get(userId, offset, maxVideoPerRequest)
		if err != nil {
			return nil, err
		}
		videos = append(videos, videosBatch...)

		if len(videosBatch) < maxVideoPerRequest {
			break
		}
		offset += maxVideoPerRequest

	}

	return videos, nil
}
