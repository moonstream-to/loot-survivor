package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
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

var TotalLeaderboardEventScores map[string]int = map[string]int{
	Event_Game_Game_DiscoveredHealth:    1,
	Event_Game_Game_DiscoveredGold:      1,
	Event_Game_Game_DiscoveredBeast:     1,
	Event_Game_Game_DodgedObstacle:      9,
	Event_Game_Game_HitByObstacle:       2,
	Event_Game_Game_AttackedBeast:       2,
	Event_Game_Game_AmbushedByBeast:     1,
	Event_Game_Game_SlayedBeast:         10,
	Event_Game_Game_FleeFailed:          2,
	Event_Game_Game_FleeSucceeded:       1,
	Event_Game_Game_PurchasedItems:      1,
	Event_Game_Game_PurchasedPotions:    1,
	Event_Game_Game_AdventurerLeveledUp: 5,
	Event_Game_Game_AdventurerUpgraded:  2,
	Event_Game_Game_IdleDeathPenalty:    -100,
	Event_Game_Game_AdventurerDied:      0,
	"MaxLevelOfBeastSlayed":             0,
}

func LootSurvivorLeaderboard(eventsFile *os.File) ([]LeaderboardScore, error) {
	// Score component -> adventurer -> value
	subscores := map[string]map[string]int{
		Event_Game_Game_DiscoveredHealth:    make(map[string]int),
		Event_Game_Game_DiscoveredGold:      make(map[string]int),
		Event_Game_Game_DiscoveredBeast:     make(map[string]int),
		Event_Game_Game_DodgedObstacle:      make(map[string]int),
		Event_Game_Game_HitByObstacle:       make(map[string]int),
		Event_Game_Game_AmbushedByBeast:     make(map[string]int),
		Event_Game_Game_SlayedBeast:         make(map[string]int),
		Event_Game_Game_FleeFailed:          make(map[string]int),
		Event_Game_Game_FleeSucceeded:       make(map[string]int),
		Event_Game_Game_PurchasedItems:      make(map[string]int),
		Event_Game_Game_PurchasedPotions:    make(map[string]int),
		Event_Game_Game_AdventurerLeveledUp: make(map[string]int),
		Event_Game_Game_AdventurerUpgraded:  make(map[string]int),
		Event_Game_Game_IdleDeathPenalty:    make(map[string]int),
		Event_Game_Game_AdventurerDied:      make(map[string]int),
		"MaxLevelOfBeastSlayed":             make(map[string]int),
	}

	names := make(map[string]string)
	activeOwners := make(map[string]string)
	scanner := bufio.NewScanner(eventsFile)
	for scanner.Scan() {
		line := scanner.Text()
		var partialEvent PartialEvent
		unmarshalErr := json.Unmarshal([]byte(line), &partialEvent)
		if unmarshalErr != nil {
			return []LeaderboardScore{}, unmarshalErr
		}

		if partialEvent.Name == Event_Game_Game_DiscoveredHealth {
			var event Game_Game_DiscoveredHealth
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.Discovery.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.Discovery.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_DiscoveredHealth][adventurer]
			subscores[Event_Game_Game_DiscoveredHealth][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_DiscoveredGold {
			var event Game_Game_DiscoveredGold
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.Discovery.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.Discovery.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_DiscoveredGold][adventurer]
			subscores[Event_Game_Game_DiscoveredGold][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_DiscoveredBeast {
			var event Game_Game_DiscoveredBeast
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner

			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_DiscoveredBeast][adventurer]
			subscores[Event_Game_Game_DiscoveredBeast][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_DodgedObstacle {
			var event Game_Game_DodgedObstacle
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.ObstacleEvent.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.ObstacleEvent.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_DodgedObstacle][adventurer]
			subscores[Event_Game_Game_DodgedObstacle][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_HitByObstacle {
			var event Game_Game_HitByObstacle
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.ObstacleEvent.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.ObstacleEvent.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_HitByObstacle][adventurer]
			subscores[Event_Game_Game_HitByObstacle][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_AmbushedByBeast {
			var event Game_Game_AmbushedByBeast
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_AmbushedByBeast][adventurer]
			subscores[Event_Game_Game_AmbushedByBeast][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_SlayedBeast {
			var event Game_Game_SlayedBeast
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_SlayedBeast][adventurer]
			subscores[Event_Game_Game_SlayedBeast][adventurer] = score + 1

			maxLevel := subscores["MaxLevelOfBeastSlayed"][adventurer]
			if int(event.BeastSpecs.Level) > maxLevel {
				subscores["MaxLevelOfBeastSlayed"][adventurer] = int(event.BeastSpecs.Level)
			}
		} else if partialEvent.Name == Event_Game_Game_FleeFailed {
			var event Game_Game_FleeFailed
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.FleeEvent.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.FleeEvent.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_FleeFailed][adventurer]
			subscores[Event_Game_Game_FleeFailed][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_FleeSucceeded {
			var event Game_Game_FleeSucceeded
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.FleeEvent.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.FleeEvent.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_FleeSucceeded][adventurer]
			subscores[Event_Game_Game_FleeSucceeded][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_PurchasedItems {
			var event Game_Game_PurchasedItems
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerStateWithBag.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerStateWithBag.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_PurchasedItems][adventurer]
			subscores[Event_Game_Game_PurchasedItems][adventurer] = score + len(event.Purchases)
		} else if partialEvent.Name == Event_Game_Game_PurchasedPotions {
			var event Game_Game_PurchasedPotions
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_PurchasedPotions][adventurer]
			subscores[Event_Game_Game_PurchasedPotions][adventurer] = score + int(event.Quantity)
		} else if partialEvent.Name == Event_Game_Game_AdventurerLeveledUp {
			var event Game_Game_AdventurerLeveledUp
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_AdventurerLeveledUp][adventurer]
			subscores[Event_Game_Game_AdventurerLeveledUp][adventurer] = score + int(event.NewLevel-event.PreviousLevel)
		} else if partialEvent.Name == Event_Game_Game_AdventurerUpgraded {
			var event Game_Game_AdventurerUpgraded
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerStateWithBag.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerStateWithBag.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_AdventurerUpgraded][adventurer]
			subscores[Event_Game_Game_AdventurerUpgraded][adventurer] = score + int(event.CharismaIncrease+event.DexterityIncrease+event.IntelligenceIncrease+event.StrengthIncrease+event.VitalityIncrease+event.WisdomIncrease)
		} else if partialEvent.Name == Event_Game_Game_IdleDeathPenalty {
			var event Game_Game_IdleDeathPenalty
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_IdleDeathPenalty][adventurer]
			subscores[Event_Game_Game_IdleDeathPenalty][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_AdventurerDied {
			var event Game_Game_AdventurerDied
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			owner := event.AdventurerState.Owner
			activeOwners[adventurer] = owner

			score := subscores[Event_Game_Game_AdventurerDied][adventurer]
			subscores[Event_Game_Game_AdventurerDied][adventurer] = score + 1
		} else if partialEvent.Name == Event_Game_Game_StartGame {
			var event Game_Game_StartGame
			unmarshalErr := json.Unmarshal(partialEvent.Event, &event)
			if unmarshalErr != nil {
				return []LeaderboardScore{}, unmarshalErr
			}

			adventurerRaw := big.NewInt(0)
			adventurerRaw.SetString(event.AdventurerState.AdventurerId, 0)
			adventurer := adventurerRaw.String()

			nameRaw := event.AdventurerMeta.Name
			name := string(nameRaw.Bytes())

			names[adventurer] = fmt.Sprintf("%s - %s", name, adventurer)
		}
	}

	scores := make(map[string]int)
	pointsData := make(map[string]map[string]interface{})
	for scoreComponent, data := range subscores {
		for adventurer, subscore := range data {
			scores[adventurer] += TotalLeaderboardEventScores[scoreComponent] * subscore
			if _, ok := pointsData[adventurer]; !ok {
				pointsData[adventurer] = make(map[string]interface{})
			}
			clean_score_component := strings.TrimPrefix(scoreComponent, "game::Game::")
			pointsData[adventurer][clean_score_component] = subscore
		}
	}
	leaderboard := make([]LeaderboardScore, len(scores))
	i := 0
	for adventurer, score := range scores {
		leaderboard[i] = LeaderboardScore{
			Address:    names[adventurer],
			Score:      int(score),
			PointsData: pointsData[adventurer],
		}
		i++
	}

	return leaderboard, nil
}
