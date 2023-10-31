package main

import (
	"errors"

	"github.com/NethermindEth/juno/core/felt"
)

/*
Events from LootSurvivor ABI:
- [x] game::Game::StartGame -- hash: 023c34c070d9c09046f7f5a319c0d6d482c1f74a5926166f6ff44e5302c4b5b3
- [x] game::Game::UpgradesAvailable -- hash: b497e78370ca3376efb8bd098ba912913a571e447c1b2c1ae4de95899d564f
- [ ] game::Game::DiscoveredHealth -- hash: 0219a5a75d3a985c139001f09646ef572f59a6a0b5f044b163d8118c311e30ba
- [ ] game::Game::DiscoveredGold -- hash: 15b7d298d615bc8a7f835a151be6f79848d7a28abba637e6608feac27255a2
- [ ] game::Game::DodgedObstacle -- hash: 033f51c3c3f7cce204e753c2254ff046ea56bfadfa22dee85814cbee84c9a63f
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

var ErrIncorrectEventKey error = errors.New("incorrect event key")
var ErrIncorrectParameters error = errors.New("incorrect parameters")
var ErrParsingAdventurerState error = errors.New("could not parse adventurer state")
var ErrParsingAdventurer error = errors.New("could not parse adventurer")
var ErrParsingAdventurerMetadata error = errors.New("could not parse adventurer metadata")
var ErrParsingStats error = errors.New("could not parse stats")
var ErrParsingCombatSpec error = errors.New("could not parse combat spec")

type ParsedEvent struct {
	Name  string `json:"name"`
	Event interface{}
}

var EVENT_UNKNOWN string = "UNKNOWN"
var EVENT_START_GAME string = "StartGame"
var EVENT_UPGRADES_AVAILABLE string = "UpgradesAvailable"
var EVENT_SLAYED_BEAST string = "SlayedBeast"

var StartGameHash string = "023c34c070d9c09046f7f5a319c0d6d482c1f74a5926166f6ff44e5302c4b5b3"
var UpgradesAvailableHash string = "b497e78370ca3376efb8bd098ba912913a571e447c1b2c1ae4de95899d564f"
var SlayedBeastHash string = "0335e768ceca00415f9ee04d58d9aebc613c76b43863445e7e33c7138184442e"

type Parser struct {
	StartGameFelt         *felt.Felt
	UpgradesAvailableFelt *felt.Felt
	SlayedBeastFelt       *felt.Felt
}

func NewParser() (*Parser, error) {
	var feltErr error
	parser := &Parser{}
	parser.StartGameFelt, feltErr = FeltFromHexString(StartGameHash)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.UpgradesAvailableFelt, feltErr = FeltFromHexString(UpgradesAvailableHash)
	if feltErr != nil {
		return parser, feltErr
	}

	parser.SlayedBeastFelt, feltErr = FeltFromHexString(SlayedBeastHash)
	if feltErr != nil {
		return parser, feltErr
	}

	return parser, nil
}

func (p *Parser) Parse(event CrawledEvent) (ParsedEvent, error) {
	defaultResult := ParsedEvent{Name: EVENT_UNKNOWN, Event: event}
	if p.IsStartGame(event) {
		parsedEvent, parseErr := p.ParseStartGame(event)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: EVENT_START_GAME, Event: parsedEvent}, nil
	} else if p.IsUpgradesAvailable(event) {
		parsedEvent, parseErr := p.ParseUpgradesAvailable(event)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: EVENT_UPGRADES_AVAILABLE, Event: parsedEvent}, nil
	} else if p.IsSlayedBeast(event) {
		parsedEvent, parseErr := p.ParseSlayedBeast(event)
		if parseErr != nil {
			return defaultResult, parseErr
		}
		return ParsedEvent{Name: EVENT_SLAYED_BEAST, Event: parsedEvent}, nil
	}
	return defaultResult, nil
}

// game::Game::StartGame -- hash: 023c34c070d9c09046f7f5a319c0d6d482c1f74a5926166f6ff44e5302c4b5b3
type StartGameEvent struct {
	CrawledEvent
	AdventurerState AdventurerState    `json:"adventurer_state"`
	AdventurerMeta  AdventurerMetadata `json:"adventurer_meta"`
	RevealBlock     uint64             `json:"reveal_block"`
}

func (p *Parser) IsStartGame(event CrawledEvent) bool {
	return p.StartGameFelt.Cmp(event.PrimaryKey) == 0
}

func (p *Parser) ParseStartGame(event CrawledEvent) (StartGameEvent, error) {
	result := StartGameEvent{
		CrawledEvent: event,
	}

	if p.StartGameFelt.Cmp(event.PrimaryKey) != 0 {
		return result, ErrIncorrectEventKey
	}

	if len(event.Parameters) != 51 {
		return result, ErrIncorrectParameters
	}

	adventurerState, adventurerStateErr := ParseAdventurerState(event.Parameters[0:41])
	if adventurerStateErr != nil {
		return result, adventurerStateErr
	}
	result.AdventurerState = adventurerState

	result.AdventurerMeta = AdventurerMetadata{
		StartBlock: event.Parameters[41].Uint64(),
		StartingStats: Stats{
			Strength:     event.Parameters[42].Uint64(),
			Dexterity:    event.Parameters[43].Uint64(),
			Vitality:     event.Parameters[44].Uint64(),
			Intelligence: event.Parameters[45].Uint64(),
			Wisdom:       event.Parameters[46].Uint64(),
			Charisma:     event.Parameters[47].Uint64(),
			Luck:         event.Parameters[48].Uint64(),
		},
		Name: event.Parameters[49].String(),
	}

	result.RevealBlock = event.Parameters[50].Uint64()

	return result, nil
}

// game::Game::UpgradesAvailable -- hash: b497e78370ca3376efb8bd098ba912913a571e447c1b2c1ae4de95899d564f
type UpgradesAvailableEvent struct {
	CrawledEvent
	AdventurerState AdventurerState `json:"adventurer_state"`
	Items           []uint64        `json:"items"`
}

func (p *Parser) IsUpgradesAvailable(event CrawledEvent) bool {
	return p.UpgradesAvailableFelt.Cmp(event.PrimaryKey) == 0
}

func (p *Parser) ParseUpgradesAvailable(event CrawledEvent) (UpgradesAvailableEvent, error) {
	result := UpgradesAvailableEvent{
		CrawledEvent: event,
	}

	if p.UpgradesAvailableFelt.Cmp(event.PrimaryKey) != 0 {
		return result, ErrIncorrectEventKey
	}

	// First 63 parameters are AdventurerState
	if len(event.Parameters) < 63 {
		return result, ErrIncorrectParameters
	}

	adventurerState, adventurerStateErr := ParseAdventurerState(event.Parameters[0:63])
	if adventurerStateErr != nil {
		return result, nil
	}
	result.AdventurerState = adventurerState

	items := []uint64{}
	if len(event.Parameters) > 63 {
		items = make([]uint64, len(event.Parameters)-63)
		for i, _ := range items {
			items[i] = event.Parameters[i+63].Uint64()
		}
	}
	result.Items = items

	return result, nil
}

// game::Game::SlayedBeast -- hash: 0335e768ceca00415f9ee04d58d9aebc613c76b43863445e7e33c7138184442e
type SlayedBeastEvent struct {
	CrawledEvent
	AdventurerState    AdventurerState `json:"adventurer_state"`
	Seed               string          `json:"seed"`
	ID                 uint64          `json:"id"`
	BeastSpecs         CombatSpec      `json:"beast_specs"`
	DamageDealt        uint64          `json:"damage_dealt"`
	CriticalHit        bool            `json:"critical_hit"`
	XPEarnedAdventurer uint64          `json:"xp_earned_adventurer"`
	XPEarnedItems      uint64          `json:"xp_earned_items"`
	GoldEarned         uint64          `json:"gold_earned"`
}

func (p *Parser) IsSlayedBeast(event CrawledEvent) bool {
	return p.SlayedBeastFelt.Cmp(event.PrimaryKey) == 0
}

func (p *Parser) ParseSlayedBeast(event CrawledEvent) (SlayedBeastEvent, error) {
	result := SlayedBeastEvent{
		CrawledEvent: event,
	}

	if p.SlayedBeastFelt.Cmp(event.PrimaryKey) != 0 {
		return result, ErrIncorrectEventKey
	}

	if len(event.Parameters) != 54 {
		return result, ErrIncorrectParameters
	}

	adventurerState, adventurerStateErr := ParseAdventurerState(event.Parameters[:41])
	if adventurerStateErr != nil {
		return result, adventurerStateErr
	}
	result.AdventurerState = adventurerState

	result.Seed = event.Parameters[41].String()
	result.ID = event.Parameters[42].Uint64()

	beastSpecs, beastSpecsErr := ParseCombatSpec(event.Parameters[43:49])
	if beastSpecsErr != nil {
		return result, beastSpecsErr
	}
	result.BeastSpecs = beastSpecs

	result.DamageDealt = event.Parameters[49].Uint64()
	result.CriticalHit = event.Parameters[50].Uint64() > 0
	result.XPEarnedAdventurer = event.Parameters[51].Uint64()
	result.XPEarnedItems = event.Parameters[52].Uint64()
	result.GoldEarned = event.Parameters[53].Uint64()

	return result, nil
}

// Core data structures used in Loot Survivor events
type AdventurerState struct {
	Owner        string     `json:"owner"`
	AdventurerID string     `json:"adventurer_id"`
	Adventurer   Adventurer `json:"adventurer"`
}

type Adventurer struct {
	LastActionBlock     uint64        `json:"last_action_block"`
	Health              uint64        `json:"health"`
	XP                  uint64        `json:"xp"`
	Stats               Stats         `json:"stats"`
	Gold                uint64        `json:"gold"`
	Weapon              ItemPrimitive `json:"weapon"`
	Chest               ItemPrimitive `json:"chest"`
	Head                ItemPrimitive `json:"head"`
	Waist               ItemPrimitive `json:"waist"`
	Foot                ItemPrimitive `json:"foot"`
	Hand                ItemPrimitive `json:"hand"`
	Neck                ItemPrimitive `json:"neck"`
	Ring                ItemPrimitive `json:"ring"`
	BeastHealth         uint64        `json:"beast_health"`
	StatPointsAvailable uint64        `json:"stat_points_available"`
	ActionsPerBlock     uint64        `json:"actions_per_block"`
	Mutated             bool          `json:"mutated"`
}

type AdventurerMetadata struct {
	StartBlock    uint64 `json:"start_block"`
	StartingStats Stats  `json:"starting_stats"`
	Name          string `json:"name"`
}

type Stats struct {
	Strength     uint64 `json:"strength"`
	Dexterity    uint64 `json:"dexterity"`
	Vitality     uint64 `json:"vitality"`
	Intelligence uint64 `json:"intelligence"`
	Wisdom       uint64 `json:"wisdom"`
	Charisma     uint64 `json:"charisma"`
	Luck         uint64 `json:"luck"`
}

type ItemPrimitive struct {
	ID       uint64 `json:"id"`
	XP       uint64 `json:"xp"`
	Metadata uint64 `json:"metadata"`
}

type CombatSpec struct {
	Tier     string        `json:"tier"`
	ItemType string        `json:"item_type"`
	Level    uint64        `json:"level"`
	Specials SpecialPowers `json:"specials"`
}

type SpecialPowers struct {
	Special1 uint64 `json:"special1"`
	Special2 uint64 `json:"special2"`
	Special3 uint64 `json:"special3"`
}

func Tier(parameter *felt.Felt) string {
	parameterInt := parameter.Uint64()
	switch parameterInt {
	case 0:
		return "None"
	case 1:
		return "T1"
	case 2:
		return "T2"
	case 3:
		return "T3"
	case 4:
		return "T4"
	case 5:
		return "T5"
	}
	return "UNKNOWN"
}

func ItemType(parameter *felt.Felt) string {
	parameterInt := parameter.Uint64()
	switch parameterInt {
	case 0:
		return "None"
	case 1:
		return "Magic_or_Cloth"
	case 2:
		return "Blade_or_Hide"
	case 3:
		return "Bludgeon_or_Metal"
	case 4:
		return "Necklace"
	case 5:
		return "Ring"
	}
	return "UNKNOWN"
}

func ParseAdventurerState(parameters []*felt.Felt) (AdventurerState, error) {
	if len(parameters) != 41 {
		return AdventurerState{}, ErrParsingAdventurerState
	}

	adventurer, adventurerErr := ParseAdventurer(parameters[2:41])
	if adventurerErr != nil {
		return AdventurerState{}, adventurerErr
	}

	return AdventurerState{
		Owner:        parameters[0].String(),
		AdventurerID: parameters[1].String(),
		Adventurer:   adventurer,
	}, nil
}

func ParseAdventurer(parameters []*felt.Felt) (Adventurer, error) {
	if len(parameters) != 39 {
		return Adventurer{}, ErrParsingAdventurer
	}
	stats, statsErr := ParseStats(parameters[3:10])
	if statsErr != nil {
		return Adventurer{}, statsErr
	}
	return Adventurer{
		LastActionBlock: parameters[0].Uint64(),
		Health:          parameters[1].Uint64(),
		XP:              parameters[2].Uint64(),
		Stats:           stats,
		Gold:            parameters[10].Uint64(),
		Weapon: ItemPrimitive{
			ID:       parameters[11].Uint64(),
			XP:       parameters[12].Uint64(),
			Metadata: parameters[13].Uint64(),
		},
		Chest: ItemPrimitive{
			ID:       parameters[14].Uint64(),
			XP:       parameters[15].Uint64(),
			Metadata: parameters[16].Uint64(),
		},
		Head: ItemPrimitive{
			ID:       parameters[17].Uint64(),
			XP:       parameters[18].Uint64(),
			Metadata: parameters[19].Uint64(),
		},
		Waist: ItemPrimitive{
			ID:       parameters[20].Uint64(),
			XP:       parameters[21].Uint64(),
			Metadata: parameters[22].Uint64(),
		},
		Foot: ItemPrimitive{
			ID:       parameters[23].Uint64(),
			XP:       parameters[24].Uint64(),
			Metadata: parameters[25].Uint64(),
		},
		Hand: ItemPrimitive{
			ID:       parameters[26].Uint64(),
			XP:       parameters[27].Uint64(),
			Metadata: parameters[28].Uint64(),
		},
		Neck: ItemPrimitive{
			ID:       parameters[29].Uint64(),
			XP:       parameters[30].Uint64(),
			Metadata: parameters[31].Uint64(),
		},
		Ring: ItemPrimitive{
			ID:       parameters[32].Uint64(),
			XP:       parameters[33].Uint64(),
			Metadata: parameters[34].Uint64(),
		},
		BeastHealth:         parameters[35].Uint64(),
		StatPointsAvailable: parameters[36].Uint64(),
		ActionsPerBlock:     parameters[37].Uint64(),
		Mutated:             parameters[38].Uint64() > 0,
	}, nil
}

func ParseAdventurerMetadata(parameters []*felt.Felt) (AdventurerMetadata, error) {
	if len(parameters) != 9 {
		return AdventurerMetadata{}, ErrParsingAdventurerMetadata
	}

	stats, statsErr := ParseStats(parameters[1:8])
	if statsErr != nil {
		return AdventurerMetadata{}, statsErr
	}

	return AdventurerMetadata{
		StartBlock:    parameters[0].Uint64(),
		StartingStats: stats,
		Name:          parameters[8].String(),
	}, nil
}

func ParseStats(parameters []*felt.Felt) (Stats, error) {
	if len(parameters) != 7 {
		return Stats{}, ErrParsingStats
	}
	return Stats{
		Strength:     parameters[0].Uint64(),
		Dexterity:    parameters[1].Uint64(),
		Vitality:     parameters[2].Uint64(),
		Intelligence: parameters[3].Uint64(),
		Wisdom:       parameters[4].Uint64(),
		Charisma:     parameters[5].Uint64(),
		Luck:         parameters[6].Uint64(),
	}, nil
}

func ParseCombatSpec(parameters []*felt.Felt) (CombatSpec, error) {
	if len(parameters) != 6 {
		return CombatSpec{}, ErrParsingCombatSpec
	}
	return CombatSpec{
		Tier:     Tier(parameters[0]),
		ItemType: ItemType(parameters[1]),
		Level:    parameters[2].Uint64(),
		Specials: SpecialPowers{
			Special1: parameters[3].Uint64(),
			Special2: parameters[4].Uint64(),
			Special3: parameters[5].Uint64(),
		},
	}, nil
}
