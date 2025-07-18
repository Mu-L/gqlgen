// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/plugin/federation/fedruntime"
	model "github.com/99designs/gqlgen/plugin/federation/testdata/computedrequires/generated/models"
)

var (
	ErrUnknownType  = errors.New("unknown type")
	ErrTypeNotFound = errors.New("type not found")
)

func (ec *executionContext) __resolve__service(ctx context.Context) (fedruntime.Service, error) {
	if ec.DisableIntrospection {
		return fedruntime.Service{}, errors.New("federated introspection disabled")
	}

	var sdl []string

	for _, src := range sources {
		if src.BuiltIn {
			continue
		}
		sdl = append(sdl, src.Input)
	}

	return fedruntime.Service{
		SDL: strings.Join(sdl, "\n"),
	}, nil
}

func (ec *executionContext) __resolve_entities(ctx context.Context, representations []map[string]any) []fedruntime.Entity {
	list := make([]fedruntime.Entity, len(representations))

	repsMap := ec.buildRepresentationGroups(ctx, representations)

	switch len(repsMap) {
	case 0:
		return list
	case 1:
		for typeName, reps := range repsMap {
			ec.resolveEntityGroup(ctx, typeName, reps, list)
		}
		return list
	default:
		var g sync.WaitGroup
		g.Add(len(repsMap))
		for typeName, reps := range repsMap {
			go func(typeName string, reps []EntityWithIndex) {
				ec.resolveEntityGroup(ctx, typeName, reps, list)
				g.Done()
			}(typeName, reps)
		}
		g.Wait()
		return list
	}
}

type EntityWithIndex struct {
	// The index in the original representation array
	index  int
	entity EntityRepresentation
}

// EntityRepresentation is the JSON representation of an entity sent by the Router
// used as the inputs for us to resolve.
//
// We make it a map because we know the top level JSON is always an object.
type EntityRepresentation map[string]any

// We group entities by typename so that we can parallelize their resolution.
// This is particularly helpful when there are entity groups in multi mode.
func (ec *executionContext) buildRepresentationGroups(
	ctx context.Context,
	representations []map[string]any,
) map[string][]EntityWithIndex {
	repsMap := make(map[string][]EntityWithIndex)
	for i, rep := range representations {
		typeName, ok := rep["__typename"].(string)
		if !ok {
			// If there is no __typename, we just skip the representation;
			// we just won't be resolving these unknown types.
			ec.Error(ctx, errors.New("__typename must be an existing string"))
			continue
		}

		repsMap[typeName] = append(repsMap[typeName], EntityWithIndex{
			index:  i,
			entity: rep,
		})
	}

	return repsMap
}

func (ec *executionContext) resolveEntityGroup(
	ctx context.Context,
	typeName string,
	reps []EntityWithIndex,
	list []fedruntime.Entity,
) {
	if isMulti(typeName) {
		err := ec.resolveManyEntities(ctx, typeName, reps, list)
		if err != nil {
			ec.Error(ctx, err)
		}
	} else {
		// if there are multiple entities to resolve, parallelize (similar to
		// graphql.FieldSet.Dispatch)
		var e sync.WaitGroup
		e.Add(len(reps))
		for i, rep := range reps {
			i, rep := i, rep
			go func(i int, rep EntityWithIndex) {
				entity, err := ec.resolveEntity(ctx, typeName, rep.entity)
				if err != nil {
					ec.Error(ctx, err)
				} else {
					list[rep.index] = entity
				}
				e.Done()
			}(i, rep)
		}
		e.Wait()
	}
}

func isMulti(typeName string) bool {
	switch typeName {
	case "MultiHello":
		return true
	case "MultiHelloMultipleRequires":
		return true
	case "MultiHelloRequires":
		return true
	case "MultiHelloWithError":
		return true
	case "MultiPlanetRequiresNested":
		return true
	default:
		return false
	}
}

func (ec *executionContext) resolveEntity(
	ctx context.Context,
	typeName string,
	rep EntityRepresentation,
) (e fedruntime.Entity, err error) {
	// we need to do our own panic handling, because we may be called in a
	// goroutine, where the usual panic handling can't catch us
	defer func() {
		if r := recover(); r != nil {
			err = ec.Recover(ctx, r)
		}
	}()

	switch typeName {
	case "Hello":
		resolverName, err := entityResolverNameForHello(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "Hello": %w`, err)
		}
		switch resolverName {

		case "findHelloByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findHelloByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindHelloByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "Hello": %w`, err)
			}

			return entity, nil
		}
	case "HelloMultiSingleKeys":
		resolverName, err := entityResolverNameForHelloMultiSingleKeys(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "HelloMultiSingleKeys": %w`, err)
		}
		switch resolverName {

		case "findHelloMultiSingleKeysByKey1AndKey2":
			id0, err := ec.unmarshalNString2string(ctx, rep["key1"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findHelloMultiSingleKeysByKey1AndKey2(): %w`, err)
			}
			id1, err := ec.unmarshalNString2string(ctx, rep["key2"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 1 for findHelloMultiSingleKeysByKey1AndKey2(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindHelloMultiSingleKeysByKey1AndKey2(ctx, id0, id1)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "HelloMultiSingleKeys": %w`, err)
			}

			return entity, nil
		}
	case "HelloWithErrors":
		resolverName, err := entityResolverNameForHelloWithErrors(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "HelloWithErrors": %w`, err)
		}
		switch resolverName {

		case "findHelloWithErrorsByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findHelloWithErrorsByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindHelloWithErrorsByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "HelloWithErrors": %w`, err)
			}

			return entity, nil
		}
	case "Person":
		resolverName, err := entityResolverNameForPerson(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "Person": %w`, err)
		}
		switch resolverName {

		case "findPersonByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findPersonByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindPersonByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "Person": %w`, err)
			}

			return entity, nil
		}
	case "PlanetMultipleRequires":
		resolverName, err := entityResolverNameForPlanetMultipleRequires(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "PlanetMultipleRequires": %w`, err)
		}
		switch resolverName {

		case "findPlanetMultipleRequiresByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findPlanetMultipleRequiresByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindPlanetMultipleRequiresByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "PlanetMultipleRequires": %w`, err)
			}

			return entity, nil
		}
	case "PlanetRequires":
		resolverName, err := entityResolverNameForPlanetRequires(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "PlanetRequires": %w`, err)
		}
		switch resolverName {

		case "findPlanetRequiresByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findPlanetRequiresByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindPlanetRequiresByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "PlanetRequires": %w`, err)
			}

			return entity, nil
		}
	case "PlanetRequiresNested":
		resolverName, err := entityResolverNameForPlanetRequiresNested(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "PlanetRequiresNested": %w`, err)
		}
		switch resolverName {

		case "findPlanetRequiresNestedByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findPlanetRequiresNestedByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindPlanetRequiresNestedByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "PlanetRequiresNested": %w`, err)
			}

			return entity, nil
		}
	case "World":
		resolverName, err := entityResolverNameForWorld(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "World": %w`, err)
		}
		switch resolverName {

		case "findWorldByHelloNameAndFoo":
			id0, err := ec.unmarshalNString2string(ctx, rep["hello"].(map[string]any)["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findWorldByHelloNameAndFoo(): %w`, err)
			}
			id1, err := ec.unmarshalNString2string(ctx, rep["foo"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 1 for findWorldByHelloNameAndFoo(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindWorldByHelloNameAndFoo(ctx, id0, id1)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "World": %w`, err)
			}

			return entity, nil
		}
	case "WorldName":
		resolverName, err := entityResolverNameForWorldName(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "WorldName": %w`, err)
		}
		switch resolverName {

		case "findWorldNameByName":
			id0, err := ec.unmarshalNString2string(ctx, rep["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findWorldNameByName(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindWorldNameByName(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "WorldName": %w`, err)
			}

			return entity, nil
		}
	case "WorldWithMultipleKeys":
		resolverName, err := entityResolverNameForWorldWithMultipleKeys(ctx, rep)
		if err != nil {
			return nil, fmt.Errorf(`finding resolver for Entity "WorldWithMultipleKeys": %w`, err)
		}
		switch resolverName {

		case "findWorldWithMultipleKeysByHelloNameAndFoo":
			id0, err := ec.unmarshalNString2string(ctx, rep["hello"].(map[string]any)["name"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findWorldWithMultipleKeysByHelloNameAndFoo(): %w`, err)
			}
			id1, err := ec.unmarshalNString2string(ctx, rep["foo"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 1 for findWorldWithMultipleKeysByHelloNameAndFoo(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindWorldWithMultipleKeysByHelloNameAndFoo(ctx, id0, id1)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "WorldWithMultipleKeys": %w`, err)
			}

			return entity, nil
		case "findWorldWithMultipleKeysByBar":
			id0, err := ec.unmarshalNInt2int(ctx, rep["bar"])
			if err != nil {
				return nil, fmt.Errorf(`unmarshalling param 0 for findWorldWithMultipleKeysByBar(): %w`, err)
			}
			entity, err := ec.resolvers.Entity().FindWorldWithMultipleKeysByBar(ctx, id0)
			if err != nil {
				return nil, fmt.Errorf(`resolving Entity "WorldWithMultipleKeys": %w`, err)
			}

			return entity, nil
		}

	}
	return nil, fmt.Errorf("%w: %s", ErrUnknownType, typeName)
}

func (ec *executionContext) resolveManyEntities(
	ctx context.Context,
	typeName string,
	reps []EntityWithIndex,
	list []fedruntime.Entity,
) (err error) {
	// we need to do our own panic handling, because we may be called in a
	// goroutine, where the usual panic handling can't catch us
	defer func() {
		if r := recover(); r != nil {
			err = ec.Recover(ctx, r)
		}
	}()

	switch typeName {

	case "MultiHello":
		resolverName, err := entityResolverNameForMultiHello(ctx, reps[0].entity)
		if err != nil {
			return fmt.Errorf(`finding resolver for Entity "MultiHello": %w`, err)
		}
		switch resolverName {

		case "findManyMultiHelloByNames":
			typedReps := make([]*model.MultiHelloByNamesInput, len(reps))

			for i, rep := range reps {
				id0, err := ec.unmarshalNString2string(ctx, rep.entity["name"])
				if err != nil {
					return errors.New(fmt.Sprintf("Field %s undefined in schema.", "name"))
				}

				typedReps[i] = &model.MultiHelloByNamesInput{
					Name: id0,
				}
			}

			entities, err := ec.resolvers.Entity().FindManyMultiHelloByNames(ctx, typedReps)
			if err != nil {
				return err
			}

			for i, entity := range entities {
				list[reps[i].index] = entity
			}
			return nil

		default:
			return fmt.Errorf("unknown resolver: %s", resolverName)
		}

	case "MultiHelloMultipleRequires":
		resolverName, err := entityResolverNameForMultiHelloMultipleRequires(ctx, reps[0].entity)
		if err != nil {
			return fmt.Errorf(`finding resolver for Entity "MultiHelloMultipleRequires": %w`, err)
		}
		switch resolverName {

		case "findManyMultiHelloMultipleRequiresByNames":
			typedReps := make([]*model.MultiHelloMultipleRequiresByNamesInput, len(reps))

			for i, rep := range reps {
				id0, err := ec.unmarshalNString2string(ctx, rep.entity["name"])
				if err != nil {
					return errors.New(fmt.Sprintf("Field %s undefined in schema.", "name"))
				}

				typedReps[i] = &model.MultiHelloMultipleRequiresByNamesInput{
					Name: id0,
				}
			}

			entities, err := ec.resolvers.Entity().FindManyMultiHelloMultipleRequiresByNames(ctx, typedReps)
			if err != nil {
				return err
			}

			for i, entity := range entities {

				list[reps[i].index] = entity
			}
			return nil

		default:
			return fmt.Errorf("unknown resolver: %s", resolverName)
		}

	case "MultiHelloRequires":
		resolverName, err := entityResolverNameForMultiHelloRequires(ctx, reps[0].entity)
		if err != nil {
			return fmt.Errorf(`finding resolver for Entity "MultiHelloRequires": %w`, err)
		}
		switch resolverName {

		case "findManyMultiHelloRequiresByNames":
			typedReps := make([]*model.MultiHelloRequiresByNamesInput, len(reps))

			for i, rep := range reps {
				id0, err := ec.unmarshalNString2string(ctx, rep.entity["name"])
				if err != nil {
					return errors.New(fmt.Sprintf("Field %s undefined in schema.", "name"))
				}

				typedReps[i] = &model.MultiHelloRequiresByNamesInput{
					Name: id0,
				}
			}

			entities, err := ec.resolvers.Entity().FindManyMultiHelloRequiresByNames(ctx, typedReps)
			if err != nil {
				return err
			}

			for i, entity := range entities {

				list[reps[i].index] = entity
			}
			return nil

		default:
			return fmt.Errorf("unknown resolver: %s", resolverName)
		}

	case "MultiHelloWithError":
		resolverName, err := entityResolverNameForMultiHelloWithError(ctx, reps[0].entity)
		if err != nil {
			return fmt.Errorf(`finding resolver for Entity "MultiHelloWithError": %w`, err)
		}
		switch resolverName {

		case "findManyMultiHelloWithErrorByNames":
			typedReps := make([]*model.MultiHelloWithErrorByNamesInput, len(reps))

			for i, rep := range reps {
				id0, err := ec.unmarshalNString2string(ctx, rep.entity["name"])
				if err != nil {
					return errors.New(fmt.Sprintf("Field %s undefined in schema.", "name"))
				}

				typedReps[i] = &model.MultiHelloWithErrorByNamesInput{
					Name: id0,
				}
			}

			entities, err := ec.resolvers.Entity().FindManyMultiHelloWithErrorByNames(ctx, typedReps)
			if err != nil {
				return err
			}

			for i, entity := range entities {
				list[reps[i].index] = entity
			}
			return nil

		default:
			return fmt.Errorf("unknown resolver: %s", resolverName)
		}

	case "MultiPlanetRequiresNested":
		resolverName, err := entityResolverNameForMultiPlanetRequiresNested(ctx, reps[0].entity)
		if err != nil {
			return fmt.Errorf(`finding resolver for Entity "MultiPlanetRequiresNested": %w`, err)
		}
		switch resolverName {

		case "findManyMultiPlanetRequiresNestedByNames":
			typedReps := make([]*model.MultiPlanetRequiresNestedByNamesInput, len(reps))

			for i, rep := range reps {
				id0, err := ec.unmarshalNString2string(ctx, rep.entity["name"])
				if err != nil {
					return errors.New(fmt.Sprintf("Field %s undefined in schema.", "name"))
				}

				typedReps[i] = &model.MultiPlanetRequiresNestedByNamesInput{
					Name: id0,
				}
			}

			entities, err := ec.resolvers.Entity().FindManyMultiPlanetRequiresNestedByNames(ctx, typedReps)
			if err != nil {
				return err
			}

			for i, entity := range entities {

				list[reps[i].index] = entity
			}
			return nil

		default:
			return fmt.Errorf("unknown resolver: %s", resolverName)
		}

	default:
		return errors.New("unknown type: " + typeName)
	}
}

func entityResolverNameForHello(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for Hello", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for Hello", ErrTypeNotFound))
			break
		}
		return "findHelloByName", nil
	}
	return "", fmt.Errorf("%w for Hello due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForHelloMultiSingleKeys(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["key1"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"key1\" for HelloMultiSingleKeys", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		m = rep
		val, ok = m["key2"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"key2\" for HelloMultiSingleKeys", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for HelloMultiSingleKeys", ErrTypeNotFound))
			break
		}
		return "findHelloMultiSingleKeysByKey1AndKey2", nil
	}
	return "", fmt.Errorf("%w for HelloMultiSingleKeys due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForHelloWithErrors(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for HelloWithErrors", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for HelloWithErrors", ErrTypeNotFound))
			break
		}
		return "findHelloWithErrorsByName", nil
	}
	return "", fmt.Errorf("%w for HelloWithErrors due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForMultiHello(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for MultiHello", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for MultiHello", ErrTypeNotFound))
			break
		}
		return "findManyMultiHelloByNames", nil
	}
	return "", fmt.Errorf("%w for MultiHello due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForMultiHelloMultipleRequires(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for MultiHelloMultipleRequires", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for MultiHelloMultipleRequires", ErrTypeNotFound))
			break
		}
		return "findManyMultiHelloMultipleRequiresByNames", nil
	}
	return "", fmt.Errorf("%w for MultiHelloMultipleRequires due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForMultiHelloRequires(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for MultiHelloRequires", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for MultiHelloRequires", ErrTypeNotFound))
			break
		}
		return "findManyMultiHelloRequiresByNames", nil
	}
	return "", fmt.Errorf("%w for MultiHelloRequires due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForMultiHelloWithError(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for MultiHelloWithError", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for MultiHelloWithError", ErrTypeNotFound))
			break
		}
		return "findManyMultiHelloWithErrorByNames", nil
	}
	return "", fmt.Errorf("%w for MultiHelloWithError due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForMultiPlanetRequiresNested(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for MultiPlanetRequiresNested", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for MultiPlanetRequiresNested", ErrTypeNotFound))
			break
		}
		return "findManyMultiPlanetRequiresNestedByNames", nil
	}
	return "", fmt.Errorf("%w for MultiPlanetRequiresNested due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForPerson(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for Person", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for Person", ErrTypeNotFound))
			break
		}
		return "findPersonByName", nil
	}
	return "", fmt.Errorf("%w for Person due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForPlanetMultipleRequires(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for PlanetMultipleRequires", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for PlanetMultipleRequires", ErrTypeNotFound))
			break
		}
		return "findPlanetMultipleRequiresByName", nil
	}
	return "", fmt.Errorf("%w for PlanetMultipleRequires due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForPlanetRequires(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for PlanetRequires", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for PlanetRequires", ErrTypeNotFound))
			break
		}
		return "findPlanetRequiresByName", nil
	}
	return "", fmt.Errorf("%w for PlanetRequires due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForPlanetRequiresNested(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for PlanetRequiresNested", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for PlanetRequiresNested", ErrTypeNotFound))
			break
		}
		return "findPlanetRequiresNestedByName", nil
	}
	return "", fmt.Errorf("%w for PlanetRequiresNested due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForWorld(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["hello"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"hello\" for World", ErrTypeNotFound))
			break
		}
		if m, ok = val.(map[string]any); !ok {
			// nested field value is not a map[string]interface so don't use it
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to nested Key Field \"hello\" value not matching map[string]any for World", ErrTypeNotFound))
			break
		}
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for World", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		m = rep
		val, ok = m["foo"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"foo\" for World", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for World", ErrTypeNotFound))
			break
		}
		return "findWorldByHelloNameAndFoo", nil
	}
	return "", fmt.Errorf("%w for World due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForWorldName(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for WorldName", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for WorldName", ErrTypeNotFound))
			break
		}
		return "findWorldNameByName", nil
	}
	return "", fmt.Errorf("%w for WorldName due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}

func entityResolverNameForWorldWithMultipleKeys(ctx context.Context, rep EntityRepresentation) (string, error) {
	// we collect errors because a later entity resolver may work fine
	// when an entity has multiple keys
	entityResolverErrs := []error{}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["hello"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"hello\" for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		if m, ok = val.(map[string]any); !ok {
			// nested field value is not a map[string]interface so don't use it
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to nested Key Field \"hello\" value not matching map[string]any for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		val, ok = m["name"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"name\" for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		m = rep
		val, ok = m["foo"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"foo\" for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		return "findWorldWithMultipleKeysByHelloNameAndFoo", nil
	}
	for {
		var (
			m   EntityRepresentation
			val any
			ok  bool
		)
		_ = val
		// if all of the KeyFields values for this resolver are null,
		// we shouldn't use use it
		allNull := true
		m = rep
		val, ok = m["bar"]
		if !ok {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to missing Key Field \"bar\" for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		if allNull {
			allNull = val == nil
		}
		if allNull {
			entityResolverErrs = append(entityResolverErrs,
				fmt.Errorf("%w due to all null value KeyFields for WorldWithMultipleKeys", ErrTypeNotFound))
			break
		}
		return "findWorldWithMultipleKeysByBar", nil
	}
	return "", fmt.Errorf("%w for WorldWithMultipleKeys due to %v", ErrTypeNotFound,
		errors.Join(entityResolverErrs...).Error())
}
