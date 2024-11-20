package orderbook

import (
	rbt "github.com/Heian0/LeGoTradingEngine/internal/utils/redblacktree"
)

type LevelMap struct {
	levelMap         *rbt.Tree
	levelMapIterator rbt.Iterator
}

func UInt64Comparator(a, b interface{}) int {
	ua, ub := a.(uint64), b.(uint64)
	if ua < ub {
		return -1
	} else if ua > ub {
		return 1
	}
	return 0
}

func NewLevelMap() *LevelMap {
	lvlMap := rbt.NewWith(UInt64Comparator)
	return &LevelMap{
		levelMap:         lvlMap,
		levelMapIterator: lvlMap.Iterator(),
	}
}

func (lvlMap *LevelMap) Put(price uint64, level *Level) {
	lvlMap.levelMap.Put(price, level)
}

func (lvlMap *LevelMap) PutWithHint(price uint64, level *Level, hint *rbt.Node) {
	lvlMap.levelMap.PutWithHint(price, level, hint)
}

func (lvlMap *LevelMap) Get(price uint64) (*Level, bool) {
	levelInterface, found := lvlMap.levelMap.Get(price)
	if !found {
		return nil, false
	}
	level, isLevelType := levelInterface.(*Level)
	if !isLevelType {
		panic("Found a non level object in LevelMap")
	}
	return level, true
}

func (lvlMap *LevelMap) Delete(price uint64) {
	lvlMap.levelMap.Remove(price)
}

func (lvlMap *LevelMap) Emplace(price uint64, levelSide Side, symbolId uint64) *Level {
	levelInterface, levelInMap := lvlMap.levelMap.Get(price)
	if levelInMap {
		level, isLevelType := levelInterface.(Level)
		if !isLevelType {
			if !lvlMap.levelMap.Empty() {
				panic("Found a non level object in LevelMap")
			}
		}
		return &level
	}
	newLevel := NewLevel(levelSide, price, symbolId)
	ptr := &newLevel
	lvlMap.Put(price, &newLevel)
	return ptr
}

func (lvlMap *LevelMap) IsEmpty() bool {
	return lvlMap.levelMap.Empty()
}

func (lvlMap *LevelMap) SetMapBegin() {
	lvlMap.levelMapIterator.Begin()
}

func (lvlMap *LevelMap) SetMapEnd() {
	lvlMap.levelMapIterator.End()
}

func (lvlMap *LevelMap) GetMapBegin() *rbt.Node {
	firstExists := lvlMap.levelMapIterator.First()
	if firstExists {
		levelNode := lvlMap.levelMapIterator.Node()
		return levelNode
	}
	return nil
}

func (lvlMap *LevelMap) GetMapEnd() *rbt.Node {
	firstExists := lvlMap.levelMapIterator.Last()
	if firstExists {
		levelNode := lvlMap.levelMapIterator.Node()
		return levelNode
	}
	return nil
}

func (lvlMap *LevelMap) EmplaceWithHint(price uint64, levelSide Side, symbolId uint64, hint *rbt.Node) *Level {
	levelInterface, levelInMap := lvlMap.levelMap.Get(price)
	if levelInMap {
		level, isLevelType := levelInterface.(Level)
		if !isLevelType {
			if !lvlMap.levelMap.Empty() {
				panic("Found a non level object in LevelMap")
			}
		}
		return &level
	}
	newLevel := NewLevel(levelSide, price, symbolId)
	ptr := &newLevel
	lvlMap.levelMap.PutWithHint(price, &newLevel, hint)
	return ptr
}

func (lvlMap *LevelMap) String() string {
	return lvlMap.levelMap.String()
}
