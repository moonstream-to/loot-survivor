package main

type SurvivorEventData struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Kind string `json:"kind"`
}

type SurvivorEvent struct {
	Name string              `json:"name"`
	Data []SurvivorEventData `json:"data"`
}

func Events(abi []map[string]interface{}) *[]SurvivorEvent {
	numEvents := 0
	for _, item := range abi {
		if item["type"] == "event" && item["kind"] == "struct" {
			numEvents++
		}
	}

	currentIndex := 0
	events := make([]SurvivorEvent, numEvents)
	for _, item := range abi {
		if item["type"] == "event" && item["kind"] == "struct" {
			events[currentIndex] = SurvivorEvent{
				Name: item["name"].(string),
			}
			if item["members"] != nil {
				membersArray := item["members"].([]interface{})
				membersMapArray := make([]map[string]interface{}, len(membersArray))
				for i, member := range membersArray {
					membersMapArray[i] = member.(map[string]interface{})
				}
				events[currentIndex].Data = make([]SurvivorEventData, len(membersArray))
				for i, member := range membersMapArray {
					events[currentIndex].Data[i] = SurvivorEventData{
						Name: member["name"].(string),
						Type: member["type"].(string),
						Kind: member["kind"].(string),
					}
				}
			}

			currentIndex++
		}
	}

	return &events
}
