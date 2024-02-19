package command

import (
	"context"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/v2/eventstore"
	"github.com/zitadel/zitadel/internal/v2/org"
	"github.com/zitadel/zitadel/internal/zerrors"
)

var (
	_ eventstore.PushReducerIntent = (*RemoveOrg)(nil)
)

type RemoveOrg struct {
	aggregate *eventstore.Aggregate
	commands  []eventstore.Command

	id       string
	sequence uint32
	state    org.State
}

func NewRemoveOrg(id string) *RemoveOrg {
	return &RemoveOrg{
		id: id,
	}
}

func (i *RemoveOrg) ToPushIntent(ctx context.Context, querier eventstore.Querier) (eventstore.PushIntent, error) {
	i.aggregate = org.NewAggregate(ctx, i.id)

	if i.id == authz.GetInstance(ctx).DefaultOrganisationID() {
		return nil, zerrors.ThrowPreconditionFailed(nil, "COMMA-wG9p1", "Errors.Org.DefaultOrgNotDeletable")
	}

	err := querier.Query(
		ctx,
		eventstore.NewFilter(
			ctx,
			eventstore.FilterEventQuery(
				eventstore.FilterAggregateTypes(org.AggregateType),
				eventstore.FilterAggregateIDs(i.id),
				eventstore.FilterEventTypes(
					org.Added.Type(),
					org.Removed.Type(),
				),
			),
		),
		i,
	)
	// TODO: check if ZITADEL project exists on this org
	if err != nil {
		return nil, err
	}

	if i.state.IsValidState(org.RemovedState) {
		// org is already removed, nothing to do
		return nil, nil
	}

	i.commands = append(i.commands, org.NewRemovedEvent(ctx))

	return i, nil
}

// Aggregate implements [eventstore.PushIntent].
func (i *RemoveOrg) Aggregate() *eventstore.Aggregate {
	return i.aggregate
}

// Commands implements [eventstore.PushIntent].
func (i *RemoveOrg) Commands() []eventstore.Command {
	return i.commands
}

// CurrentSequence implements [eventstore.PushIntent].
func (i *RemoveOrg) CurrentSequence() eventstore.CurrentSequence {
	return eventstore.SequenceAtLeast(i.sequence)
}

// Reduce implements [eventstore.Reducer].
func (i *RemoveOrg) Reduce(events ...eventstore.Event) error {
	for _, event := range events {
		switch event.Type() {
		case org.Added.Type():
			i.state = org.ActiveState
		case org.Removed.Type():
			i.state = org.RemovedState
		}
	}

	return nil
}