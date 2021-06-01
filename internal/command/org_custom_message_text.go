package command

import (
	"context"

	"golang.org/x/text/language"

	"github.com/caos/zitadel/internal/domain"
	caos_errs "github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/eventstore"
	"github.com/caos/zitadel/internal/repository/org"
)

func (c *Commands) SetOrgMessageText(ctx context.Context, resourceOwner string, mailText *domain.CustomMessageText) (*domain.ObjectDetails, error) {
	if resourceOwner == "" {
		return nil, caos_errs.ThrowInvalidArgument(nil, "ORG-2biiR", "Errors.ResourceOwnerMissing")
	}
	iamAgg := org.NewAggregate(resourceOwner, resourceOwner)
	events, existingMailText, err := c.setOrgMessageText(ctx, &iamAgg.Aggregate, mailText)
	if err != nil {
		return nil, err
	}
	pushedEvents, err := c.eventstore.PushEvents(ctx, events...)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingMailText, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToObjectDetails(&existingMailText.WriteModel), nil
}

func (c *Commands) setOrgMessageText(ctx context.Context, orgAgg *eventstore.Aggregate, message *domain.CustomMessageText) ([]eventstore.EventPusher, *OrgCustomMessageTextReadModel, error) {
	if !message.IsValid() {
		return nil, nil, caos_errs.ThrowInvalidArgument(nil, "ORG-2jfsf", "Errors.CustomText.Invalid")
	}

	existingMailText, err := c.orgCustomMessageTextWriteModelByID(ctx, orgAgg.ID, message.MessageTextType, message.Language)
	if err != nil {
		return nil, nil, err
	}
	events := make([]eventstore.EventPusher, 0)
	if existingMailText.Greeting != message.Greeting {
		if message.Greeting != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageGreeting, message.Greeting, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageGreeting, message.Language))
		}
	}
	if existingMailText.Subject != message.Subject {
		if message.Subject != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageSubject, message.Subject, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageSubject, message.Language))
		}
	}
	if existingMailText.Title != message.Title {
		if message.Title != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageTitle, message.Title, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageTitle, message.Language))
		}
	}
	if existingMailText.PreHeader != message.PreHeader {
		if message.PreHeader != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessagePreHeader, message.PreHeader, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessagePreHeader, message.Language))
		}
	}
	if existingMailText.Text != message.Text {
		if message.Text != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageText, message.Text, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageText, message.Language))
		}
	}
	if existingMailText.ButtonText != message.ButtonText {
		if message.ButtonText != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageButtonText, message.ButtonText, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageButtonText, message.Language))
		}
	}
	if existingMailText.FooterText != message.FooterText {
		if message.FooterText != "" {
			events = append(events, org.NewCustomTextSetEvent(ctx, orgAgg, message.MessageTextType, domain.MessageFooterText, message.FooterText, message.Language))
		} else {
			events = append(events, org.NewCustomTextRemovedEvent(ctx, orgAgg, message.MessageTextType, domain.MessageFooterText, message.Language))
		}
	}
	return events, existingMailText, nil
}

func (c *Commands) RemoveOrgMessageTexts(ctx context.Context, resourceOwner, mailTextType string, lang language.Tag) error {
	if resourceOwner == "" {
		return caos_errs.ThrowInvalidArgument(nil, "Org-3mfsf", "Errors.ResourceOwnerMissing")
	}
	if mailTextType == "" || lang == language.Und {
		return caos_errs.ThrowInvalidArgument(nil, "Org-j59f", "Errors.CustomMailText.Invalid")
	}
	customText, err := c.orgCustomMessageTextWriteModelByID(ctx, resourceOwner, mailTextType, lang)
	if err != nil {
		return err
	}
	if customText.State == domain.PolicyStateUnspecified || customText.State == domain.PolicyStateRemoved {
		return caos_errs.ThrowNotFound(nil, "Org-3b8Jf", "Errors.CustomMailText.NotFound")
	}
	orgAgg := OrgAggregateFromWriteModel(&customText.WriteModel)
	_, err = c.eventstore.PushEvents(ctx, org.NewCustomTextTemplateRemovedEvent(ctx, orgAgg, mailTextType, lang))
	return err
}

func (c *Commands) orgCustomMessageTextWriteModelByID(ctx context.Context, orgID, messageType string, lang language.Tag) (*OrgCustomMessageTextReadModel, error) {
	writeModel := NewOrgCustomMessageTextWriteModel(orgID, messageType, lang)
	err := c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}
