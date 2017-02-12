package vkc

import (
	"fmt"
	"strconv"
)

const (
	maxAudioPerRequest = 5000
)

type Audio struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type audioService struct {
	client *VkClient
}

type audioGetResponse struct {
	Count int
	Items []*Audio
}

func (s *audioService) Get(userId, offset, count int) ([]*Audio, error) {
	if count > maxAudioPerRequest {
		return nil, fmt.Errorf("count should be <= %d", maxAudioPerRequest)
	}

	params := map[string]string{
		"owner_id": strconv.Itoa(userId),
		"offset":   strconv.Itoa(offset),
		"count":    strconv.Itoa(count),
	}
	req, err := s.client.NewRequest("audio.get", params)
	if err != nil {
		return nil, err
	}

	var response audioGetResponse
	_, err = s.client.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (s *audioService) GetAll(userId int) ([]*Audio, error) {
	offset := 0
	audio := make([]*Audio, 0)

	for {
		usersBatch, err := s.Get(userId, offset, maxAudioPerRequest)
		if err != nil {
			return nil, err
		}
		audio = append(audio, usersBatch...)

		if len(usersBatch) < maxAudioPerRequest {
			break
		}
		offset += maxAudioPerRequest
	}

	return audio, nil
}
