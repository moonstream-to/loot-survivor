package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
		var parsedEvent ParsedEvent
		unmarshalErr := json.Unmarshal([]byte(line), &parsedEvent)
		if unmarshalErr != nil {
			return []LeaderboardScore{}, unmarshalErr
		}

		if parsedEvent.Name == EVENT_SLAYED_BEAST {
			// Need to figure out a better way of doing this.
			jsonBytes, marshalErr := json.Marshal(parsedEvent.Event)
			if marshalErr != nil {
				return []LeaderboardScore{}, marshalErr
			}
			var event SlayedBeastEvent
			unmarshalErr := json.Unmarshal(jsonBytes, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.AdventurerState.AdventurerID

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := scores[adventurer]
			scores[adventurer] = score + 1

			maxLevel := maxLevels[adventurer]
			if event.BeastSpecs.Level > maxLevel {
				maxLevels[adventurer] = event.BeastSpecs.Level
			}
		} else if parsedEvent.Name == EVENT_START_GAME {
			// Need to figure out a better way of doing this.
			jsonBytes, marshalErr := json.Marshal(parsedEvent.Event)
			if marshalErr != nil {
				return []LeaderboardScore{}, marshalErr
			}
			var event StartGameEvent
			unmarshalErr := json.Unmarshal(jsonBytes, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurer := event.AdventurerState.AdventurerID
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
