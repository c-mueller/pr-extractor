package extractor

import (
	"go.mongodb.org/mongo-driver/bson"
)

func (e *Extractor) loadPullRequests() error {
	filter := map[string]string{
		"type": "PullRequestEvent",
	}
	return e.runDataFetcher(filter, "events", func(data bson.Raw) error {
		var evt PullRequestEvent
		_ = bson.Unmarshal(data, &evt)

		return e.insertPullRequest(evt, data)
	}, "pull_request_fetcher")
}

func (e *Extractor) insertPullRequest(evt PullRequestEvent, elem bson.Raw) error {
	eventId := getEventId(evt)

	prId := getPullRequestId(evt)

	var comp []byte

	if e.Config.IncludeRaw {
		comp, _ = GzipCompress(elem)
	}

	resultEvt := PullRequest{
		EventDbId:                 eventId,
		PullRequestId:             prId,
		RepoName:                  evt.Repo.Name,
		RepoUrl:                   evt.Repo.Url,
		PRUrl:                     evt.Payload.PullRequest.URL,
		PRNumber:                  evt.Payload.Number,
		State:                     evt.Payload.PullRequest.State,
		PRAuthorLogin:             evt.Payload.PullRequest.User.Login,
		PRAuthorType:              evt.Payload.PullRequest.User.Type,
		PullRequestCreatedAt:      evt.Payload.PullRequest.CreatedAt,
		PullRequestUpdatedAt:      evt.Payload.PullRequest.UpdatedAt,
		PullRequestClosedAt:       evt.Payload.PullRequest.ClosedAt,
		PullRequestMergedAt:       evt.Payload.PullRequest.MergedAt,
		EventInitiatorLogin:       evt.Actor.Login,
		EventInitiatorDisplayName: evt.Actor.DisplayLogin,
		Comments:                  evt.Payload.PullRequest.Comments,
		Commits:                   evt.Payload.PullRequest.Commits,
		Additions:                 evt.Payload.PullRequest.Additions,
		Deletions:                 evt.Payload.PullRequest.Deletions,
		FilesChanged:              evt.Payload.PullRequest.FilesChanged,
		EventTimestamp:            evt.CreatedAt,
		EventAction:               evt.Payload.Action,
		RawPayload:                comp,
	}
	return e.sqlDb.Save(&resultEvt).Error
}
