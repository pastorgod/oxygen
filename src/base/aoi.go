package base

import (
	. "logger"
)

// Area Of Interest

const DISTANCE = 10

type Position struct {
	X int32
	Y int32
}

type IAOIEntity interface {

	// unique key of enity.
	Key() string

	//	OnEnter()

	//	OnLeave()
}

type AOIEntitySet struct {
	sets map[string]IAOIEntity
}

func NewAOIEntitySet() *AOIEntitySet {
	return &AOIEntitySet{
		sets: make(map[string]IAOIEntity, 2),
	}
}

func (this *AOIEntitySet) Add(object IAOIEntity) {
	if _, ok := this.sets[object.Key()]; ok {
		panic("AOIEntitySet.Add " + object.Key())
		return
	}

	this.sets[object.Key()] = object
}

func (this *AOIEntitySet) Remove(object IAOIEntity) {
	if _, ok := this.sets[object.Key()]; !ok {
		panic("AOIEntitySet.Remove " + object.Key())
		return
	}

	delete(this.sets, object.Key())
}

func (this *AOIEntitySet) Each(handler func(IAOIEntity)) {
	for _, value := range this.sets {
		handler(value)
	}
}

type IAOINotifier interface {

	// 目标进入这个位置
	OnEnter(object IAOIEntity, pos Position)

	// 目标离开这个位置
	OnLeave(object IAOIEntity, pos Position)

	// 目标位置发生改变
	OnChange(object IAOIEntity, oldPos, newPos Position)
}

type IAOISpace interface {
	// init aoi module.
	AOIInit(notifier IAOINotifier, cell_rect, max_rect Position)
	// on enity enter.
	AOIEnter(object IAOIEntity, pos Position)
	// on enity leave.
	AOILeave(object IAOIEntity, pos Position)
	// on enity update pos.
	AOIUpdate(object IAOIEntity, oldPos, newPos Position)
}

type CellAOI struct {
	notifier  IAOINotifier
	cell_rect Position
	max_rect  Position
	edge_rect Position
	cells     [][]*AOIEntitySet
}

func NewCellAOI() *CellAOI {
	return &CellAOI{}
}

func (this *CellAOI) AOIInit(notifier IAOINotifier, cell_rect, max_rect Position) {
	Assert(cell_rect.X > 0 && cell_rect.Y > 0, "invalid cell_rect")
	Assert(cell_rect.X <= max_rect.X && cell_rect.Y <= max_rect.Y, "invalid max_rect")

	this.notifier = notifier
	this.cell_rect, this.max_rect = cell_rect, max_rect

	// 计算 M * N 格子
	m, n := max_rect.X/cell_rect.X, max_rect.Y/cell_rect.Y

	// 计算最大格
	this.edge_rect.X, this.edge_rect.Y = m-1, n-1

	// 初始化格子
	this.cells = make([][]*AOIEntitySet, m)
	for i := int32(0); i < m; i++ {
		this.cells[i] = make([]*AOIEntitySet, n)
		for j := int32(0); j < n; j++ {
			this.cells[i][j] = NewAOIEntitySet()
		}
	}

	DEBUG("CellAOI.Init: [%d-%d] [%d-%d] => [%d-%d]", cell_rect.X, cell_rect.Y, max_rect.X, max_rect.Y, m, n)
}

// 转换为对应的格子
func (this *CellAOI) transCellPos(pos Position) (dst_cell Position) {
	dst_cell.X = MinInt32(this.edge_rect.X, FloorInt32(float32(pos.X)/float32(this.cell_rect.X)))
	dst_cell.Y = MinInt32(this.edge_rect.Y, FloorInt32(float32(pos.Y)/float32(this.cell_rect.Y)))
	return
}

// 检查是否在格子里
func (this *CellAOI) isInRect(pos, start, end Position) bool {
	return pos.X >= start.X && pos.X <= end.X && pos.Y >= start.Y && pos.Y <= end.Y
}

func (this *CellAOI) getPosLimit(pos Position, r int32, max Position) (start, end Position) {

	if pos.X-r < 0 {
		start.X = 0
		end.X = 2 * r
	} else if pos.X+r > max.X {
		start.X = max.X - 2*r
		end.X = max.X
	} else {
		start.X = pos.X - r
		end.X = pos.X + r
	}

	if pos.Y-r < 0 {
		start.Y = 0
		end.Y = 2 * r
	} else if pos.Y+r > max.Y {
		start.Y = max.Y - 2*r
		end.Y = max.Y
	} else {
		start.Y = pos.Y - r
		end.Y = pos.Y + r
	}

	if start.X < 0 {
		start.X = 0
	}

	if end.X > max.Y {
		end.X = max.Y
	}

	if start.Y < 0 {
		start.Y = 0
	}

	if end.Y > max.Y {
		end.Y = max.Y
	}

	return
}

func (this *CellAOI) ForEachCellEntity(cell Position, r int32, handler func(IAOIEntity, Position)) {
	this.listCells(cell, r, func(x, y int32) {
		this.cells[x][y].Each(func(object IAOIEntity) {
			handler(object, Position{X: x, Y: y})
		})
	})
}

func (this *CellAOI) listCells(cell Position, r int32, handler func(x, y int32)) {

	start, end := this.getPosLimit(cell, r, this.edge_rect)

	for i := start.X; i < end.X; i++ {
		for j := start.Y; j < end.Y; j++ {
			handler(i, j)
		}
	}
}

func (this *CellAOI) AOIEnter(object IAOIEntity, pos Position) {
	pos = this.transCellPos(pos)
	this.cells[pos.X][pos.Y].Add(object)
	// 加入格子
	this.notifier.OnEnter(object, pos)
}

func (this *CellAOI) AOILeave(object IAOIEntity, pos Position) {
	pos = this.transCellPos(pos)
	this.cells[pos.X][pos.Y].Remove(object)
	// 离开格子
	this.notifier.OnLeave(object, pos)
}

func (this *CellAOI) AOIUpdate(object IAOIEntity, oldPos, newPos Position) {

	p1 := this.transCellPos(oldPos)
	p2 := this.transCellPos(newPos)

	// 木有移动出格子
	if p1.X == p2.X && p1.Y == p2.Y {
		return
	}

	// 从原来的地方移除掉
	this.cells[p1.X][p1.Y].Remove(object)

	// 加入新的格子
	this.cells[p2.X][p2.Y].Add(object)

	// 切换格子了
	this.notifier.OnChange(object, p1, p2)

	return
}
