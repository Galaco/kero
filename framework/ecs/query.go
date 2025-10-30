package ecs

import (
	"fmt"
	"sort"
	"strings"
)

// Query finds entities with specific component combinations.
// Queries are built using a fluent API and can be cached for performance.
type Query struct {
	with    []ComponentType
	without []ComponentType
	world   *World
	key     string // Cache key
}

// QueryBuilder provides a fluent API for building queries
type QueryBuilder struct {
	query *Query
}

// Query creates a new query builder
func (w *World) Query() *QueryBuilder {
	return &QueryBuilder{
		query: &Query{
			world:   w,
			with:    make([]ComponentType, 0, 4),
			without: make([]ComponentType, 0, 4),
		},
	}
}

// With adds a required component type to the query
func (qb *QueryBuilder) With(componentType ComponentType) *QueryBuilder {
	qb.query.with = append(qb.query.with, componentType)
	return qb
}

// Without adds an excluded component type to the query
func (qb *QueryBuilder) Without(componentType ComponentType) *QueryBuilder {
	qb.query.without = append(qb.query.without, componentType)
	return qb
}

// Build finalizes the query
func (qb *QueryBuilder) Build() *Query {
	// Generate cache key
	qb.query.key = generateQueryKey(qb.query.with, qb.query.without)
	return qb.query
}

// Entities returns all entities matching the query.
// This is the primary method for iterating entities in systems.
func (q *Query) Entities() []Entity {
	// Check cache
	q.world.queryCacheMu.RLock()
	if cached, exists := q.world.queryCache[q.key]; exists {
		q.world.queryCacheMu.RUnlock()
		return cached.execute()
	}
	q.world.queryCacheMu.RUnlock()

	// Cache miss - execute and cache
	q.world.queryCacheMu.Lock()
	q.world.queryCache[q.key] = q
	q.world.queryCacheMu.Unlock()

	return q.execute()
}

// execute runs the query and returns matching entities
func (q *Query) execute() []Entity {
	if len(q.with) == 0 {
		return []Entity{} // No required components = no results
	}

	// Start with smallest component storage for efficiency
	var smallestStorage interface{}
	var smallestSize int = int(^uint(0) >> 1) // Max int

	for _, componentType := range q.with {
		storage := q.world.getStorageByType(componentType)
		if storage == nil {
			return []Entity{} // Component type has no entities
		}

		if sizer, ok := storage.(interface{ Len() int }); ok {
			size := sizer.Len()
			if size < smallestSize {
				smallestSize = size
				smallestStorage = storage
			}
		}
	}

	if smallestStorage == nil {
		return []Entity{}
	}

	// Get entities from smallest storage
	var entities []Entity
	if entityGetter, ok := smallestStorage.(interface{ GetEntities() []Entity }); ok {
		entities = entityGetter.GetEntities()
	} else {
		return []Entity{}
	}

	// Filter entities that have all required components and none of the excluded ones
	result := make([]Entity, 0, len(entities))
	for _, entity := range entities {
		if q.matches(entity) {
			result = append(result, entity)
		}
	}

	return result
}

// matches checks if an entity matches the query criteria
func (q *Query) matches(entity Entity) bool {
	// Check all "with" components
	for _, componentType := range q.with {
		storage := q.world.getStorageByType(componentType)
		if storage == nil {
			return false
		}

		if hasChecker, ok := storage.(interface{ Has(Entity) bool }); ok {
			if !hasChecker.Has(entity) {
				return false
			}
		} else {
			return false
		}
	}

	// Check all "without" components
	for _, componentType := range q.without {
		storage := q.world.getStorageByType(componentType)
		if storage == nil {
			continue // No storage = entity doesn't have it = good
		}

		if hasChecker, ok := storage.(interface{ Has(Entity) bool }); ok {
			if hasChecker.Has(entity) {
				return false // Entity has excluded component
			}
		}
	}

	return true
}

// generateQueryKey creates a unique string key for caching queries
func generateQueryKey(with, without []ComponentType) string {
	var sb strings.Builder

	// Sort for consistent keys
	withCopy := make([]ComponentType, len(with))
	copy(withCopy, with)
	sort.Slice(withCopy, func(i, j int) bool {
		return withCopy[i] < withCopy[j]
	})

	withoutCopy := make([]ComponentType, len(without))
	copy(withoutCopy, without)
	sort.Slice(withoutCopy, func(i, j int) bool {
		return withoutCopy[i] < withoutCopy[j]
	})

	sb.WriteString("with:")
	for i, ct := range withCopy {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", ct))
	}

	if len(withoutCopy) > 0 {
		sb.WriteString("|without:")
		for i, ct := range withoutCopy {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("%d", ct))
		}
	}

	return sb.String()
}

// Count returns the number of entities matching the query
func (q *Query) Count() int {
	return len(q.Entities())
}

// First returns the first entity matching the query, or NullEntity if none found
func (q *Query) First() Entity {
	entities := q.Entities()
	if len(entities) == 0 {
		return NullEntity
	}
	return entities[0]
}

// Each iterates over all matching entities and calls the provided function
func (q *Query) Each(fn func(Entity)) {
	for _, entity := range q.Entities() {
		fn(entity)
	}
}
