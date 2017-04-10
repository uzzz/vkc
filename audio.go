package vkc

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
)

const (
	maxAudioPerRequest = 5000
)

type Audio struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

type audioResponse struct {
	Artist  string `json:"artist"`
	Title   string `json:"title"`
	OwnerId int    `json:"owner_id"`
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

var (
	getAudios = `
var user_ids = [{{ range $i, $e := $ -}}{{if $i}}, {{end}}{{ $e }}{{ end }}];
var data = [];
var i = 0;
while (i < user_ids.length) {
    var response = API.audio.get({
		"owner_id": user_ids[i],
		"count": 6000,
		"need_user": 0
	});
	if (response){
		data.push(response);
	}
	i = i + 1;
};
return data[0].items;`
	getAudiosTemplate = template.Must(template.New("get.audios").Parse(getAudios))
)

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

func (s *audioService) GetBatch(userIds []int) (map[int][]*Audio, error) {
	if len(userIds) > MaximumExecuteBatchSize {
		return nil, fmt.Errorf("batch size can't be more than %d", MaximumExecuteBatchSize)
	}

	results := make(map[int][]*Audio)

	var buf bytes.Buffer
	err := getAudiosTemplate.Execute(&buf, userIds)
	if err != nil {
		return nil, err
	}

	var audios []*audioResponse

	err = s.client.Execute(buf.String(), &audios)
	if err != nil {
		return nil, err
	}

	for _, respItem := range audios {
		results[respItem.OwnerId] = append(results[respItem.OwnerId], &Audio{
			Artist: respItem.Artist,
			Title:  respItem.Title,
		})
	}

	return results, nil
}
