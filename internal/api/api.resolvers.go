package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apigen "atomys.codes/stud42/internal/api/generated"
	typesgen "atomys.codes/stud42/internal/api/generated/types"
	"atomys.codes/stud42/internal/discord"
	modelsutils "atomys.codes/stud42/internal/models"
	"atomys.codes/stud42/internal/models/generated"
	"atomys.codes/stud42/internal/models/generated/account"
	"atomys.codes/stud42/internal/models/generated/campus"
	"atomys.codes/stud42/internal/models/generated/location"
	"atomys.codes/stud42/internal/models/generated/user"
	"atomys.codes/stud42/internal/models/gotype"
	"atomys.codes/stud42/internal/pkg/searchengine"
	"atomys.codes/stud42/pkg/utils"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (r *mutationResolver) CreateFriendship(ctx context.Context, userID uuid.UUID) (bool, error) {
	cu, err := CurrentUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	if userID == cu.ID {
		return false, fmt.Errorf("cannot befriend yourself")
	}

	if _, err := r.client.User.UpdateOne(cu).AddFollowingIDs(userID).Save(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) DeleteFriendship(ctx context.Context, userID uuid.UUID) (bool, error) {
	cu, err := CurrentUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	if _, err := r.client.User.UpdateOne(cu).RemoveFollowingIDs(userID).Save(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) UpdateSettings(ctx context.Context, input typesgen.SettingsInput) (*gotype.Settings, error) {
	cu, err := CurrentUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedUser, err := r.client.User.UpdateOne(cu).SetSettings(gotype.Settings(input)).Save(ctx)
	if err != nil {
		return nil, err
	}
	return &updatedUser.Settings, nil
}

func (r *mutationResolver) InternalCreateUser(ctx context.Context, input typesgen.CreateUserInput) (uuid.UUID, error) {
	campusID, err := r.client.Campus.Query().Where(campus.DuoID(input.CurrentDuoCampusID)).FirstID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("cannot find campus")
		return uuid.Nil, err
	}

	return r.client.User.Create().
		SetEmail(input.Email).
		SetDuoLogin(input.DuoLogin).
		SetDuoID(input.DuoID).
		SetFirstName(input.FirstName).
		SetNillableUsualFirstName(input.UsualFirstName).
		SetLastName(input.LastName).
		SetNillablePoolYear(input.PoolYear).
		SetNillablePoolMonth(input.PoolMonth).
		SetNillablePhone(input.Phone).
		SetNillableDuoAvatarURL(input.DuoAvatarURL).
		SetNillableDuoAvatarSmallURL(input.DuoAvatarSmallURL).
		SetIsStaff(input.IsStaff).
		SetNillableCurrentCampusID(&campusID).
		SetIsAUser(true).
		OnConflictColumns(user.FieldDuoID).
		UpdateNewValues().
		ID(ctx)
}

func (r *mutationResolver) InternalLinkAccount(ctx context.Context, input typesgen.LinkAccountInput) (*generated.Account, error) {
	id, err := r.client.Account.Create().
		SetProvider(input.Provider.String()).
		SetProviderAccountID(input.ProviderAccountID).
		SetUsername(input.Username).
		SetType(input.Type.String()).
		SetAccessToken(input.AccessToken).
		SetNillableRefreshToken(input.RefreshToken).
		SetTokenType(input.TokenType).
		SetNillableExpiresAt(input.ExpiresAt).
		SetScope(input.Scope).
		SetUserID(uuid.MustParse(input.UserID)).
		OnConflictColumns(account.FieldProvider, account.FieldProviderAccountID).
		UpdateNewValues().
		ID(ctx)

	if err != nil {
		return nil, err
	}

	account, err := r.client.Account.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	go accountLinkCallback(ctx, r.client, account)

	return account, nil
}

func (r *mutationResolver) InviteOnDiscord(ctx context.Context) (bool, error) {
	cu, err := CurrentUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	acc, err := r.client.Account.Query().Where(account.UserID(cu.ID), account.Provider(string(typesgen.ProviderDiscord))).Only(ctx)
	if err != nil {
		return false, err
	}

	err = discord.DefaultClient().InviteOnOurDiscord(ctx, acc.AccessToken, acc.ProviderAccountID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *queryResolver) Me(ctx context.Context) (*generated.User, error) {
	return CurrentUserFromContext(ctx)
}

func (r *queryResolver) SearchUser(ctx context.Context, query string, onlyOnline *bool) ([]*generated.User, error) {
	usersID, err := searchengine.NewClient().SearchUser(query, *onlyOnline)
	if err != nil {
		return nil, err
	}

	return r.client.User.Query().Modify(func(s *sql.Selector) {
		s.
			Select("users.*").
			SetDistinct(false).
			FromExpr(sql.Expr("unnest($1::uuid[]) WITH ORDINALITY AS x(id, order_nr)", "{"+strings.Join(utils.StringifySlice(usersID), ", ")+"}")).
			Join(sql.Table(user.Table).As(user.Table)).
			On(
				sql.Table(user.Table).C(user.FieldID),
				sql.Table("x").C(user.FieldID),
			).
			OrderExpr(sql.Expr("x.order_nr"))
	}).All(ctx)
}

func (r *queryResolver) Campus(ctx context.Context, id uuid.UUID) (*generated.Campus, error) {
	return r.client.Campus.Query().Where(campus.ID(id)).Only(ctx)
}

func (r *queryResolver) User(ctx context.Context, id uuid.UUID) (*generated.User, error) {
	return r.client.User.Query().Where(user.ID(id)).Only(ctx)
}

func (r *queryResolver) Location(ctx context.Context, id uuid.UUID) (*generated.Location, error) {
	return r.client.Location.Query().Where(location.ID(id)).Only(ctx)
}

func (r *queryResolver) Locations(ctx context.Context, page typesgen.PageInput, campusID uuid.UUID) (*generated.LocationConnection, error) {
	return r.client.Location.Query().
		Where(location.CampusID(campusID)).
		WithCampus().
		WithUser().
		Where(location.EndAtIsNil()).
		Paginate(ctx, page.After, &page.First, page.Before, page.Last)
}

func (r *queryResolver) LocationsByCampusName(ctx context.Context, page typesgen.PageInput, campusName string) (*generated.LocationConnection, error) {
	return r.client.Campus.Query().
		Where(campus.Name(campusName)).
		QueryLocations().
		WithCampus().
		WithUser().
		Where(location.EndAtIsNil()).
		Paginate(ctx, page.After, &page.First, page.Before, page.Last)
}

func (r *queryResolver) LocationsByCluster(ctx context.Context, page typesgen.PageInput, campusName string, identifierPrefix *string) (*generated.LocationConnection, error) {
	return r.client.Campus.Query().
		Where(campus.NameEqualFold(campusName)).
		QueryLocations().
		WithCampus().
		WithUser().
		Where(location.IdentifierHasPrefix(*identifierPrefix), location.EndAtIsNil()).
		Paginate(ctx, page.After, &page.First, page.Before, page.Last)
}

func (r *queryResolver) LocationsStatsByPrefixes(ctx context.Context, campusName string, identifierPrefixes []string) ([]*typesgen.LocationStats, error) {
	sqlResults := []*typesgen.LocationStats{}
	prefixes := make([]any, len(identifierPrefixes))
	identifierMaxSize := 2

	// We need to convert the prefixes to any to be able to use them in the query
	// and we also need to know the max size of the prefixes to be able to
	// know how many characters we need to take from the identifier to filter
	// after the request.
	for i, prefix := range identifierPrefixes {
		// Validate the length of the prefix.
		if len(prefix) < 2 || len(prefix) > 4 {
			return nil, fmt.Errorf("invalid prefix size. Must be between 2 and 4")
		}

		if len(prefix) > identifierMaxSize {
			identifierMaxSize = len(prefix)
		}
		prefixes[i] = prefix
	}

	err := r.client.Location.Query().
		Modify(func(s *sql.Selector) {
			s.Select(
				sql.As(sql.Table(campus.Table).C(campus.FieldID), "campusID"),
				sql.As(sql.Count(sql.Table(location.Table).C(location.FieldID)), "occupiedWorkspace"),
				sql.As(fmt.Sprintf("left(%s, %d)", sql.Table(location.Table).C(location.FieldIdentifier), identifierMaxSize), "prefix"),
			).
				Join(sql.Table(campus.Table).As(campus.Table)).
				On(
					sql.Table(campus.Table).C(campus.FieldID),
					sql.Table(location.Table).C(location.FieldCampusID),
				).
				Where(
					sql.And(
						sql.EqualFold(sql.Table(campus.Table).C(campus.FieldName), campusName),
						sql.IsNull(sql.Table(location.Table).C(location.FieldEndAt)),
						sql.ExprP(
							fmt.Sprintf(
								"%q.%q ~ $2",
								location.Table,
								location.FieldIdentifier,
							),
							"^"+strings.Join(identifierPrefixes, "|"),
						),
					),
				).
				GroupBy("(campus.id, prefix)")
		}).
		Scan(ctx, &sqlResults)

	// We need to loop over the result to calculate the total number of workspaces
	// for each prefix. This is because we can't use group by sql instruction
	// with the count function over a regex filter.
	// Need more research on this topic. Maybe there is a better way to do this.
	var finalLocationStats = make([]*typesgen.LocationStats, len(identifierPrefixes))
	for i, prefix := range identifierPrefixes {
		finalLocationStats[i] = &typesgen.LocationStats{
			Prefix: prefix,
		}

		for _, locationStat := range sqlResults {
			if finalLocationStats[i].CampusID == uuid.Nil {
				finalLocationStats[i].CampusID = locationStat.CampusID
			}

			if strings.HasPrefix(locationStat.Prefix, prefix) {
				finalLocationStats[i].OccupiedWorkspace += locationStat.OccupiedWorkspace
			}
		}
	}

	return finalLocationStats, err
}

func (r *queryResolver) MyFollowing(ctx context.Context) ([]*generated.User, error) {
	cu, _ := CurrentUserFromContext(ctx)

	withCampus := func(lq *generated.LocationQuery) {
		lq.WithCampus()
	}

	return r.client.User.Query().
		Where(user.ID(cu.ID)).
		QueryFollowing().
		WithCurrentLocation(withCampus).
		WithLastLocation(withCampus).
		// Unique is necessary because the query builder always add a DISTINCT clause
		// and cannot order the query properly by location identifier
		Unique(false).
		Order(func(s *sql.Selector) {
			//: Hack to order the friends as A -> Z over the connected status
			t := sql.Table(location.Table)
			s.LeftJoin(t).On(s.C(user.FieldCurrentLocationID), t.C(location.FieldID))
			s.OrderBy(t.C(location.FieldUserDuoLogin), s.C(user.FieldDuoLogin))
			//: Hack to order the friends as A -> Z over the connected status
		}).
		All(ctx)
}

func (r *queryResolver) InternalGetUserByAccount(ctx context.Context, provider typesgen.Provider, uid string) (*generated.User, error) {
	return r.client.Account.Query().
		Where(account.Provider(provider.String()), account.ProviderAccountID(uid)).
		QueryUser().
		Only(ctx)
}

func (r *queryResolver) InternalGetUserByEmail(ctx context.Context, email string) (*generated.User, error) {
	return r.client.User.Query().
		Where(user.Email(email)).
		Only(ctx)
}

func (r *queryResolver) InternalGetUser(ctx context.Context, id uuid.UUID) (*generated.User, error) {
	return r.client.User.Get(ctx, id)
}

func (r *userResolver) IsSwimmer(ctx context.Context, obj *generated.User) (bool, error) {
	if obj.PoolYear == nil || obj.PoolMonth == nil {
		return false, nil
	}

	now := time.Now()
	return (*obj.PoolYear == strconv.Itoa(now.Year()) &&
		strings.EqualFold(*obj.PoolMonth, now.Format("January"))), nil
}

func (r *userResolver) IsMe(ctx context.Context, obj *generated.User) (bool, error) {
	cu, _ := CurrentUserFromContext(ctx)

	return cu.ID == obj.ID, nil
}

func (r *userResolver) Flags(ctx context.Context, obj *generated.User) ([]typesgen.Flag, error) {
	return modelsutils.TranslateFlagFromORM(obj.FlagsList), nil
}

func (r *userResolver) IsFollowing(ctx context.Context, obj *generated.User) (bool, error) {
	cu, _ := CurrentUserFromContext(ctx)

	for _, f := range cu.Edges.Following {
		if f.ID == obj.ID {
			return true, nil
		}
	}

	return false, nil
}

func (r *userResolver) IsFollower(ctx context.Context, obj *generated.User) (bool, error) {
	cu, _ := CurrentUserFromContext(ctx)

	for _, f := range cu.Edges.Followers {
		if f.ID == obj.ID {
			return true, nil
		}
	}

	return false, nil
}

// Mutation returns apigen.MutationResolver implementation.
func (r *Resolver) Mutation() apigen.MutationResolver { return &mutationResolver{r} }

// Query returns apigen.QueryResolver implementation.
func (r *Resolver) Query() apigen.QueryResolver { return &queryResolver{r} }

// User returns apigen.UserResolver implementation.
func (r *Resolver) User() apigen.UserResolver { return &userResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
