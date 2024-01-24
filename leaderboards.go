package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/moonstream-to/loot-survivor/bindings"
)

var LEADERBOARDS_API_URL string = "https://engineapi.moonstream.to/leaderboard/%s/scores"

type LeaderboardScore struct {
	Address    string                 `json:"address"`
	Score      int                    `json:"score"`
	PointsData map[string]interface{} `json:"points_data"`
}

func Push(leaderboardID, accessToken string, leaderboard []LeaderboardScore, overwrite bool) error {
	leaderboardURL := fmt.Sprintf(LEADERBOARDS_API_URL, leaderboardID)
	u, parseErr := url.Parse(leaderboardURL)
	if parseErr != nil {
		return parseErr
	}
	queryParams := url.Values{}
	queryParams.Set("normalize_addresses", "false")
	if overwrite {
		queryParams.Set("overwrite", "true")
	} else {
		queryParams.Set("overwrite", "false")
	}
	u.RawQuery = queryParams.Encode()

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encodeErr := encoder.Encode(leaderboard)
	if encodeErr != nil {
		return encodeErr
	}

	req, setupErr := http.NewRequest("PUT", u.String(), &buf)
	if setupErr != nil {
		return setupErr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	httpClient := &http.Client{}
	resp, apiErr := httpClient.Do(req)
	if apiErr != nil {
		return apiErr
	}
	defer resp.Body.Close()

	fmt.Fprintf(os.Stderr, "Status: %d\n", resp.StatusCode)
	return nil
}

func BeastSlayersLeaderboard(eventsFile *os.File) ([]LeaderboardScore, error) {
	names := make(map[string]string)
	activeOwners := make(map[string]string)
	scores := make(map[string]uint64)
	maxLevels := make(map[string]uint64)
	scanner := bufio.NewScanner(eventsFile)
	for scanner.Scan() {
		line := scanner.Text()
		var partialEvent PartialEvent
		unmarshalErr := json.Unmarshal([]byte(line), &partialEvent)
		if unmarshalErr != nil {
			return []LeaderboardScore{}, unmarshalErr
		}

		if partialEvent.Name == bindings.Event_Game_Game_SlayedBeast {
			var event bindings.Game_Game_SlayedBeast
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.AdventurerState.AdventurerId

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := scores[adventurer]
			scores[adventurer] = score + 1

			maxLevel := maxLevels[adventurer]
			if event.BeastSpecs.Level > maxLevel {
				maxLevels[adventurer] = event.BeastSpecs.Level
			}
		} else if partialEvent.Name == bindings.Event_Game_Game_StartGame {
			var event bindings.Game_Game_StartGame
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.AdventurerState.AdventurerId
			name := event.AdventurerMeta.Name

			names[adventurer] = fmt.Sprintf("%s - %s", name, adventurer)
		}
	}

	leaderboard := make([]LeaderboardScore, len(scores))
	i := 0
	for adventurer, score := range scores {
		leaderboard[i] = LeaderboardScore{
			Address: names[adventurer],
			Score:   int(score),
			PointsData: map[string]interface{}{
				"max_level":    maxLevels[adventurer],
				"active_owner": activeOwners[adventurer],
			},
		}
		i++
	}

	return leaderboard, nil
}

func ArtfulDodgersLeaderboard(eventsFile *os.File) ([]LeaderboardScore, error) {
	names := make(map[string]string)
	activeOwners := make(map[string]string)
	scores := make(map[string]uint64)
	maxLevels := make(map[string]uint64)
	scanner := bufio.NewScanner(eventsFile)
	for scanner.Scan() {
		line := scanner.Text()
		var partialEvent PartialEvent
		unmarshalErr := json.Unmarshal([]byte(line), &partialEvent)
		if unmarshalErr != nil {
			return []LeaderboardScore{}, unmarshalErr
		}

		if partialEvent.Name == bindings.Event_Game_Game_DodgedObstacle {
			var event bindings.Game_Game_DodgedObstacle
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.ObstacleEvent.AdventurerState.AdventurerId

			owner := event.ObstacleEvent.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := scores[adventurer]
			scores[adventurer] = score + 1

			maxLevel := maxLevels[adventurer]
			if event.ObstacleEvent.ObstacleDetails.Level > maxLevel {
				maxLevels[adventurer] = event.ObstacleEvent.ObstacleDetails.Level
			}
		} else if partialEvent.Name == bindings.Event_Game_Game_StartGame {
			var event bindings.Game_Game_StartGame
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.AdventurerState.AdventurerId
			name := event.AdventurerMeta.Name

			names[adventurer] = fmt.Sprintf("%s - %s", name, adventurer)
		}
	}

	leaderboard := make([]LeaderboardScore, len(scores))
	i := 0
	for adventurer, score := range scores {
		leaderboard[i] = LeaderboardScore{
			Address: names[adventurer],
			Score:   int(score),
			PointsData: map[string]interface{}{
				"max_level":    maxLevels[adventurer],
				"active_owner": activeOwners[adventurer],
			},
		}
		i++
	}

	return leaderboard, nil
}
