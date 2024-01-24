package main

import (
	"encoding/json"
	"fmt"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/moonstream-to/loot-survivor/bindings"
)

/*
Events from LootSurvivor ABI:
- [x] game::Game::StartGame -- hash: 023c34c070d9c09046f7f5a319c0d6d482c1f74a5926166f6ff44e5302c4b5b3
- [x] game::Game::UpgradesAvailable -- hash: b497e78370ca3376efb8bd098ba912913a571e447c1b2c1ae4de95899d564f
- [ ] game::Game::DiscoveredHealth -- hash: 0219a5a75d3a985c139001f09646ef572f59a6a0b5f044b163d8118c311e30ba
- [ ] game::Game::DiscoveredGold -- hash: 15b7d298d615bc8a7f835a151be6f79848d7a28abba637e6608feac27255a2
- [x] game::Game::DodgedObstacle -- hash: 033f51c3c3f7cce204e753c2254ff046ea56bfadfa22dee85814cbee84c9a63f
- [ ] game::Game::HitByObstacle -- hash: 9ebf5c9567959039f24c635b67d3a000eabf12ba43906df38b9cdbd501317a
- [ ] game::Game::AmbushedByBeast -- hash: 0317b067327eb6721c6e2e3e299b24a39b58d33771aa154e03a53f8978aa7e68
- [ ] game::Game::DiscoveredBeast -- hash: 0253a02657b86e566e59683306f5d7b2032b7814d140eeab9b85f8fcca6780fb
- [ ] game::Game::AttackedBeast -- hash: 036637c9d2490697d67734e20c4923e77319f2cc0eb9c1f85139473b893882b0
- [ ] game::Game::AttackedByBeast -- hash: 03a441cdc9a394fdfd2ca0864d82abfec7108bb007924694b87428acfde5a1b6
- [x] game::Game::SlayedBeast -- hash: 0335e768ceca00415f9ee04d58d9aebc613c76b43863445e7e33c7138184442e
- [ ] game::Game::FleeFailed -- hash: 03733166aa57fb199c903b0764828bbef285fc337dec553f0796741cec15c589
- [ ] game::Game::FleeSucceeded -- hash: 02db50d7ded17829196848c38d96ff4423f8524e315d16982d55df1b3473fca6
- [ ] game::Game::AdventurerLeveledUp -- hash: 016b747f083a5bc0eb62f1465891cc9b6ce061c09e77556f949348af4e7a608a
- [ ] game::Game::PurchasedItems -- hash: 0127e97eb6a3822aeea793175c23d1dab98c510ce54a317e530ac3891cc076fd
- [ ] game::Game::PurchasedPotions -- hash: 015a142c8e70a4bcecff9338b254db8b1240d0dda34c8163e098bb5b550f2ccf
- [ ] game::Game::AdventurerUpgraded -- hash: 02158c8175c2574c95b9d6f44940661609d36927c128204bbc48f8854faae048
- [ ] game::Game::EquippedItems -- hash: 02fe099b5d05c4f7cd492117d7e640f7aedc86b509965bb35de682531fb05452
- [ ] game::Game::DroppedItems -- hash: 035694c9186ae97e1d000975d8cc4e1eefd9393affc6655cf1706bcde28540de
- [ ] game::Game::ItemsLeveledUp -- hash: cc7cceeb77459c7412ece4747120d08750a7f9cd83b643f54926fba96521fd
- [ ] game::Game::AdventurerDied -- hash: ac0d9cf65432fd092269cbdb9901ccbffeb652bd7781b362ba30b79090fc38
- [ ] game::Game::NewHighScore -- hash: 02326b9588750a7ce7c31809060a0123a05a60ae7eaa478d5cb01f3f797cb216
- [ ] game::Game::IdleDeathPenalty -- hash: 010677066134d8347763db8e41a6bd207d841c2d6728539617c4ca9d84528319
- [ ] game::Game::RewardDistribution -- hash: 02945361ab13c4fbed57e3e6c2330d6bdd49335c8f123424a35cee08c7011747
- [ ] game::Game::GameEntropyRotatedEvent -- hash: 0310b5616bf40dc2dfd971ebd30938841b7f2d57fb3e6274c1a59416e1e8bc02
- [ ] game::Game::PriceChangeEvent -- hash: 6a20da49efcea3fa37b370e38c87bedbda1d418a041009aa29faea92d7499a

Events from Beasts ABI:
- [ ] LootSurvivorBeasts::beasts::Beasts::Transfer -- hash: 99cd8bde557814842a3121e8ddfd433a539b8c9f14bf31ebf108d12e6196e9
- [ ] LootSurvivorBeasts::beasts::Beasts::Approval -- hash: 0134692b230b9e1ffa39098904722134159652b09c5bc41d88d6698779d228ff
- [ ] LootSurvivorBeasts::beasts::Beasts::ApprovalForAll -- hash: 06ad9ed7b6318f1bcffefe19df9aeb40d22c36bed567e1925a5ccde0536edd
*/

var EVENT_UNKNOWN = "UNKNOWN"

type ParsedEvent struct {
	Name  string      `json:"name"`
	Event interface{} `json:"event"`
}

type PartialEvent struct {
	Name  string          `json:"name"`
	Event json.RawMessage `json:"event"`
}

type LootSurvivorEventParser struct {
	StartGameFelt          *felt.Felt
	UpgradesAvailableFelt  *felt.Felt
	SlayedBeastFelt        *felt.Felt
	DodgedObstacleFelt     *felt.Felt
	RewardDistributionFelt *felt.Felt
}

func NewLootSurvivorParser() (*LootSurvivorEventParser, error) {
	var feltErr error
	parser := &LootSurvivorEventParser{}
	parser.StartGameFelt, feltErr = FeltFromHexString(bindings.Hash_Game_Game_StartGame)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.UpgradesAvailableFelt, feltErr = FeltFromHexString(bindings.Hash_Game_Game_UpgradesAvailable)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.SlayedBeastFelt, feltErr = FeltFromHexString(bindings.Hash_Game_Game_SlayedBeast)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.DodgedObstacleFelt, feltErr = FeltFromHexString(bindings.Hash_Game_Game_DodgedObstacle)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.RewardDistributionFelt, feltErr = FeltFromHexString(bindings.Hash_Game_Game_RewardDistribution)
	if feltErr != nil {
		return parser, feltErr
	}

	return parser, nil
}

func (p *LootSurvivorEventParser) Parse(event CrawledEvent) (ParsedEvent, error) {
	defaultResult := ParsedEvent{Name: EVENT_UNKNOWN, Event: event}
	if p.StartGameFelt.Cmp(event.PrimaryKey) == 0 {
		parsedEvent, _, parseErr := bindings.ParseGame_Game_StartGame(event.Parameters)
		if parseErr != nil {
			fmt.Println(parseErr.Error())
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: bindings.Event_Game_Game_StartGame, Event: parsedEvent}, nil
	} else if p.UpgradesAvailableFelt.Cmp(event.PrimaryKey) == 0 {
		parsedEvent, _, parseErr := bindings.ParseGame_Game_UpgradesAvailable(event.Parameters)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: bindings.Event_Game_Game_UpgradesAvailable, Event: parsedEvent}, nil
	} else if p.SlayedBeastFelt.Cmp(event.PrimaryKey) == 0 {
		parsedEvent, _, parseErr := bindings.ParseGame_Game_SlayedBeast(event.Parameters)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: bindings.Event_Game_Game_SlayedBeast, Event: parsedEvent}, nil
	} else if p.DodgedObstacleFelt.Cmp(event.PrimaryKey) == 0 {
		parsedEvent, _, parseErr := bindings.ParseGame_Game_DodgedObstacle(event.Parameters)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: bindings.Event_Game_Game_DodgedObstacle, Event: parsedEvent}, nil
	} else if p.RewardDistributionFelt.Cmp(event.PrimaryKey) == 0 {
		parsedEvent, _, parseErr := bindings.ParseGame_Game_RewardDistribution(event.Parameters)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: bindings.Event_Game_Game_RewardDistribution, Event: parsedEvent}, nil
	}
	return defaultResult, nil
}
