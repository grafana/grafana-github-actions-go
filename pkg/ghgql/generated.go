// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package ghgql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
)

// __getMilestonedPullRequestsInput is used internally by genqlient
type __getMilestonedPullRequestsInput struct {
	Owner           string `json:"owner"`
	Repo            string `json:"repo"`
	MilestoneNumber int    `json:"milestoneNumber"`
	Cursor          string `json:"cursor"`
}

// GetOwner returns __getMilestonedPullRequestsInput.Owner, and is useful for accessing the field via an interface.
func (v *__getMilestonedPullRequestsInput) GetOwner() string { return v.Owner }

// GetRepo returns __getMilestonedPullRequestsInput.Repo, and is useful for accessing the field via an interface.
func (v *__getMilestonedPullRequestsInput) GetRepo() string { return v.Repo }

// GetMilestoneNumber returns __getMilestonedPullRequestsInput.MilestoneNumber, and is useful for accessing the field via an interface.
func (v *__getMilestonedPullRequestsInput) GetMilestoneNumber() int { return v.MilestoneNumber }

// GetCursor returns __getMilestonedPullRequestsInput.Cursor, and is useful for accessing the field via an interface.
func (v *__getMilestonedPullRequestsInput) GetCursor() string { return v.Cursor }

// __getMilestonesWithTitleInput is used internally by genqlient
type __getMilestonesWithTitleInput struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Title string `json:"title"`
}

// GetOwner returns __getMilestonesWithTitleInput.Owner, and is useful for accessing the field via an interface.
func (v *__getMilestonesWithTitleInput) GetOwner() string { return v.Owner }

// GetRepo returns __getMilestonesWithTitleInput.Repo, and is useful for accessing the field via an interface.
func (v *__getMilestonesWithTitleInput) GetRepo() string { return v.Repo }

// GetTitle returns __getMilestonesWithTitleInput.Title, and is useful for accessing the field via an interface.
func (v *__getMilestonesWithTitleInput) GetTitle() string { return v.Title }

// getMilestonedPullRequestsRepository includes the requested fields of the GraphQL type Repository.
// The GraphQL type's documentation follows.
//
// A repository contains the content for a project.
type getMilestonedPullRequestsRepository struct {
	// Returns a single milestone from the current repository by number.
	Milestone getMilestonedPullRequestsRepositoryMilestone `json:"milestone"`
}

// GetMilestone returns getMilestonedPullRequestsRepository.Milestone, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepository) GetMilestone() getMilestonedPullRequestsRepositoryMilestone {
	return v.Milestone
}

// getMilestonedPullRequestsRepositoryMilestone includes the requested fields of the GraphQL type Milestone.
// The GraphQL type's documentation follows.
//
// Represents a Milestone object on a given repository.
type getMilestonedPullRequestsRepositoryMilestone struct {
	// A list of pull requests associated with the milestone.
	PullRequests getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection `json:"pullRequests"`
}

// GetPullRequests returns getMilestonedPullRequestsRepositoryMilestone.PullRequests, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestone) GetPullRequests() getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection {
	return v.PullRequests
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection includes the requested fields of the GraphQL type PullRequestConnection.
// The GraphQL type's documentation follows.
//
// The connection type for PullRequest.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection struct {
	// Information to aid in pagination.
	PageInfo getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo `json:"pageInfo"`
	// A list of nodes.
	Nodes []getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest `json:"nodes"`
}

// GetPageInfo returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection.PageInfo, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection) GetPageInfo() getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo {
	return v.PageInfo
}

// GetNodes returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection.Nodes, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnection) GetNodes() []getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest {
	return v.Nodes
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest includes the requested fields of the GraphQL type PullRequest.
// The GraphQL type's documentation follows.
//
// A repository pull request.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest struct {
	// Identifies the pull request number.
	Number int `json:"number"`
	// Identifies the pull request title.
	Title string `json:"title"`
	// The body as Markdown.
	Body string `json:"body"`
	// A list of labels associated with the object.
	Labels getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection `json:"labels"`
	// The actor who authored the comment.
	Author getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor `json:"-"`
	// Identifies the name of the head Ref associated with the pull request, even if the ref has been deleted.
	HeadRefName string `json:"headRefName"`
}

// GetNumber returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Number, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetNumber() int {
	return v.Number
}

// GetTitle returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Title, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetTitle() string {
	return v.Title
}

// GetBody returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Body, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetBody() string {
	return v.Body
}

// GetLabels returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Labels, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetLabels() getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection {
	return v.Labels
}

// GetAuthor returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Author, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetAuthor() getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor {
	return v.Author
}

// GetHeadRefName returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.HeadRefName, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) GetHeadRefName() string {
	return v.HeadRefName
}

func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) UnmarshalJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	var firstPass struct {
		*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest
		Author json.RawMessage `json:"author"`
		graphql.NoUnmarshalJSON
	}
	firstPass.getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest = v

	err := json.Unmarshal(b, &firstPass)
	if err != nil {
		return err
	}

	{
		dst := &v.Author
		src := firstPass.Author
		if len(src) != 0 && string(src) != "null" {
			err = __unmarshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor(
				src, dst)
			if err != nil {
				return fmt.Errorf(
					"unable to unmarshal getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Author: %w", err)
			}
		}
	}
	return nil
}

type __premarshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest struct {
	Number int `json:"number"`

	Title string `json:"title"`

	Body string `json:"body"`

	Labels getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection `json:"labels"`

	Author json.RawMessage `json:"author"`

	HeadRefName string `json:"headRefName"`
}

func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) MarshalJSON() ([]byte, error) {
	premarshaled, err := v.__premarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(premarshaled)
}

func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest) __premarshalJSON() (*__premarshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest, error) {
	var retval __premarshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest

	retval.Number = v.Number
	retval.Title = v.Title
	retval.Body = v.Body
	retval.Labels = v.Labels
	{

		dst := &retval.Author
		src := v.Author
		var err error
		*dst, err = __marshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor(
			&src)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to marshal getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequest.Author: %w", err)
		}
	}
	retval.HeadRefName = v.HeadRefName
	return &retval, nil
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor includes the requested fields of the GraphQL interface Actor.
//
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor is implemented by the following types:
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization
// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser
// The GraphQL type's documentation follows.
//
// Represents an object which can take actions on GitHub. Typically a User or Bot.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor interface {
	implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor()
	// GetTypename returns the receiver's concrete GraphQL type-name (see interface doc for possible values).
	GetTypename() string
	// GetResourcePath returns the interface-field "resourcePath" from its implementation.
	// The GraphQL interface field's documentation follows.
	//
	// The HTTP path for this actor.
	GetResourcePath() string
	// GetLogin returns the interface-field "login" from its implementation.
	// The GraphQL interface field's documentation follows.
	//
	// The username of the actor.
	GetLogin() string
}

func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot) implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor() {
}
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount) implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor() {
}
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin) implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor() {
}
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization) implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor() {
}
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser) implementsGraphQLInterfacegetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor() {
}

func __unmarshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor(b []byte, v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor) error {
	if string(b) == "null" {
		return nil
	}

	var tn struct {
		TypeName string `json:"__typename"`
	}
	err := json.Unmarshal(b, &tn)
	if err != nil {
		return err
	}

	switch tn.TypeName {
	case "Bot":
		*v = new(getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot)
		return json.Unmarshal(b, *v)
	case "EnterpriseUserAccount":
		*v = new(getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount)
		return json.Unmarshal(b, *v)
	case "Mannequin":
		*v = new(getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin)
		return json.Unmarshal(b, *v)
	case "Organization":
		*v = new(getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization)
		return json.Unmarshal(b, *v)
	case "User":
		*v = new(getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser)
		return json.Unmarshal(b, *v)
	case "":
		return fmt.Errorf(
			"response was missing Actor.__typename")
	default:
		return fmt.Errorf(
			`unexpected concrete type for getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor: "%v"`, tn.TypeName)
	}
}

func __marshalgetMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor(v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor) ([]byte, error) {

	var typename string
	switch v := (*v).(type) {
	case *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot:
		typename = "Bot"

		result := struct {
			TypeName string `json:"__typename"`
			*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot
		}{typename, v}
		return json.Marshal(result)
	case *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount:
		typename = "EnterpriseUserAccount"

		result := struct {
			TypeName string `json:"__typename"`
			*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount
		}{typename, v}
		return json.Marshal(result)
	case *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin:
		typename = "Mannequin"

		result := struct {
			TypeName string `json:"__typename"`
			*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin
		}{typename, v}
		return json.Marshal(result)
	case *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization:
		typename = "Organization"

		result := struct {
			TypeName string `json:"__typename"`
			*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization
		}{typename, v}
		return json.Marshal(result)
	case *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser:
		typename = "User"

		result := struct {
			TypeName string `json:"__typename"`
			*getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser
		}{typename, v}
		return json.Marshal(result)
	case nil:
		return []byte("null"), nil
	default:
		return nil, fmt.Errorf(
			`unexpected concrete type for getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorActor: "%T"`, v)
	}
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot includes the requested fields of the GraphQL type Bot.
// The GraphQL type's documentation follows.
//
// A special type of user which takes actions on behalf of GitHub Apps.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot struct {
	Typename string `json:"__typename"`
	// The HTTP path for this actor.
	ResourcePath string `json:"resourcePath"`
	// The username of the actor.
	Login string `json:"login"`
}

// GetTypename returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot.Typename, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot) GetTypename() string {
	return v.Typename
}

// GetResourcePath returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot.ResourcePath, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot) GetResourcePath() string {
	return v.ResourcePath
}

// GetLogin returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot.Login, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorBot) GetLogin() string {
	return v.Login
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount includes the requested fields of the GraphQL type EnterpriseUserAccount.
// The GraphQL type's documentation follows.
//
// An account for a user who is an admin of an enterprise or a member of an enterprise through one or more organizations.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount struct {
	Typename string `json:"__typename"`
	// The HTTP path for this actor.
	ResourcePath string `json:"resourcePath"`
	// The username of the actor.
	Login string `json:"login"`
}

// GetTypename returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount.Typename, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount) GetTypename() string {
	return v.Typename
}

// GetResourcePath returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount.ResourcePath, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount) GetResourcePath() string {
	return v.ResourcePath
}

// GetLogin returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount.Login, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorEnterpriseUserAccount) GetLogin() string {
	return v.Login
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin includes the requested fields of the GraphQL type Mannequin.
// The GraphQL type's documentation follows.
//
// A placeholder user for attribution of imported data on GitHub.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin struct {
	Typename string `json:"__typename"`
	// The HTTP path for this actor.
	ResourcePath string `json:"resourcePath"`
	// The username of the actor.
	Login string `json:"login"`
}

// GetTypename returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin.Typename, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin) GetTypename() string {
	return v.Typename
}

// GetResourcePath returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin.ResourcePath, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin) GetResourcePath() string {
	return v.ResourcePath
}

// GetLogin returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin.Login, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorMannequin) GetLogin() string {
	return v.Login
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization includes the requested fields of the GraphQL type Organization.
// The GraphQL type's documentation follows.
//
// An account on GitHub, with one or more owners, that has repositories, members and teams.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization struct {
	Typename string `json:"__typename"`
	// The HTTP path for this actor.
	ResourcePath string `json:"resourcePath"`
	// The username of the actor.
	Login string `json:"login"`
}

// GetTypename returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization.Typename, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization) GetTypename() string {
	return v.Typename
}

// GetResourcePath returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization.ResourcePath, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization) GetResourcePath() string {
	return v.ResourcePath
}

// GetLogin returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization.Login, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorOrganization) GetLogin() string {
	return v.Login
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser includes the requested fields of the GraphQL type User.
// The GraphQL type's documentation follows.
//
// A user is an individual's account on GitHub that owns repositories and can make new content.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser struct {
	Typename string `json:"__typename"`
	// The HTTP path for this actor.
	ResourcePath string `json:"resourcePath"`
	// The username of the actor.
	Login string `json:"login"`
}

// GetTypename returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser.Typename, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser) GetTypename() string {
	return v.Typename
}

// GetResourcePath returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser.ResourcePath, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser) GetResourcePath() string {
	return v.ResourcePath
}

// GetLogin returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser.Login, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestAuthorUser) GetLogin() string {
	return v.Login
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection includes the requested fields of the GraphQL type LabelConnection.
// The GraphQL type's documentation follows.
//
// The connection type for Label.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection struct {
	// A list of nodes.
	Nodes []getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel `json:"nodes"`
}

// GetNodes returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection.Nodes, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnection) GetNodes() []getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel {
	return v.Nodes
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel includes the requested fields of the GraphQL type Label.
// The GraphQL type's documentation follows.
//
// A label for categorizing Issues, Pull Requests, Milestones, or Discussions with a given Repository.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel struct {
	// Identifies the label name.
	Name string `json:"name"`
}

// GetName returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel.Name, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionNodesPullRequestLabelsLabelConnectionNodesLabel) GetName() string {
	return v.Name
}

// getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo includes the requested fields of the GraphQL type PageInfo.
// The GraphQL type's documentation follows.
//
// Information about pagination in a connection.
type getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo struct {
	// When paginating forwards, the cursor to continue.
	EndCursor string `json:"endCursor"`
	// When paginating forwards, are there more items?
	HasNextPage bool `json:"hasNextPage"`
}

// GetEndCursor returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo.EndCursor, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo) GetEndCursor() string {
	return v.EndCursor
}

// GetHasNextPage returns getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo.HasNextPage, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsRepositoryMilestonePullRequestsPullRequestConnectionPageInfo) GetHasNextPage() bool {
	return v.HasNextPage
}

// getMilestonedPullRequestsResponse is returned by getMilestonedPullRequests on success.
type getMilestonedPullRequestsResponse struct {
	// Lookup a given repository by the owner and repository name.
	Repository getMilestonedPullRequestsRepository `json:"repository"`
}

// GetRepository returns getMilestonedPullRequestsResponse.Repository, and is useful for accessing the field via an interface.
func (v *getMilestonedPullRequestsResponse) GetRepository() getMilestonedPullRequestsRepository {
	return v.Repository
}

// getMilestonesWithTitleRepository includes the requested fields of the GraphQL type Repository.
// The GraphQL type's documentation follows.
//
// A repository contains the content for a project.
type getMilestonesWithTitleRepository struct {
	// A list of milestones associated with the repository.
	Milestones getMilestonesWithTitleRepositoryMilestonesMilestoneConnection `json:"milestones"`
}

// GetMilestones returns getMilestonesWithTitleRepository.Milestones, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepository) GetMilestones() getMilestonesWithTitleRepositoryMilestonesMilestoneConnection {
	return v.Milestones
}

// getMilestonesWithTitleRepositoryMilestonesMilestoneConnection includes the requested fields of the GraphQL type MilestoneConnection.
// The GraphQL type's documentation follows.
//
// The connection type for Milestone.
type getMilestonesWithTitleRepositoryMilestonesMilestoneConnection struct {
	// A list of nodes.
	Nodes []getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone `json:"nodes"`
}

// GetNodes returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnection.Nodes, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnection) GetNodes() []getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone {
	return v.Nodes
}

// getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone includes the requested fields of the GraphQL type Milestone.
// The GraphQL type's documentation follows.
//
// Represents a Milestone object on a given repository.
type getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone struct {
	// Identifies the number of the milestone.
	Number int    `json:"number"`
	Id     string `json:"id"`
	// Indicates if the object is closed (definition of closed may depend on type)
	Closed bool `json:"closed"`
	// Identifies the title of the milestone.
	Title string `json:"title"`
	// Identifies the date and time when the object was closed.
	ClosedAt time.Time `json:"closedAt"`
	// Identifies the due date of the milestone.
	DueOn time.Time `json:"dueOn"`
}

// GetNumber returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.Number, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetNumber() int {
	return v.Number
}

// GetId returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.Id, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetId() string {
	return v.Id
}

// GetClosed returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.Closed, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetClosed() bool {
	return v.Closed
}

// GetTitle returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.Title, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetTitle() string {
	return v.Title
}

// GetClosedAt returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.ClosedAt, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetClosedAt() time.Time {
	return v.ClosedAt
}

// GetDueOn returns getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone.DueOn, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleRepositoryMilestonesMilestoneConnectionNodesMilestone) GetDueOn() time.Time {
	return v.DueOn
}

// getMilestonesWithTitleResponse is returned by getMilestonesWithTitle on success.
type getMilestonesWithTitleResponse struct {
	// Lookup a given repository by the owner and repository name.
	Repository getMilestonesWithTitleRepository `json:"repository"`
}

// GetRepository returns getMilestonesWithTitleResponse.Repository, and is useful for accessing the field via an interface.
func (v *getMilestonesWithTitleResponse) GetRepository() getMilestonesWithTitleRepository {
	return v.Repository
}

// The query or mutation executed by getMilestonedPullRequests.
const getMilestonedPullRequests_Operation = `
query getMilestonedPullRequests ($owner: String!, $repo: String!, $milestoneNumber: Int!, $cursor: String!) {
	repository(owner: $owner, name: $repo) {
		milestone(number: $milestoneNumber) {
			pullRequests(first: 20, states: [MERGED], labels: ["add to changelog"], after: $cursor) {
				pageInfo {
					endCursor
					hasNextPage
				}
				nodes {
					number
					title
					body
					labels(first: 20) {
						nodes {
							name
						}
					}
					author {
						__typename
						resourcePath
						login
					}
					headRefName
				}
			}
		}
	}
}
`

func getMilestonedPullRequests(
	ctx context.Context,
	client graphql.Client,
	owner string,
	repo string,
	milestoneNumber int,
	cursor string,
) (*getMilestonedPullRequestsResponse, error) {
	req := &graphql.Request{
		OpName: "getMilestonedPullRequests",
		Query:  getMilestonedPullRequests_Operation,
		Variables: &__getMilestonedPullRequestsInput{
			Owner:           owner,
			Repo:            repo,
			MilestoneNumber: milestoneNumber,
			Cursor:          cursor,
		},
	}
	var err error

	var data getMilestonedPullRequestsResponse
	resp := &graphql.Response{Data: &data}

	err = client.MakeRequest(
		ctx,
		req,
		resp,
	)

	return &data, err
}

// The query or mutation executed by getMilestonesWithTitle.
const getMilestonesWithTitle_Operation = `
query getMilestonesWithTitle ($owner: String!, $repo: String!, $title: String!) {
	repository(owner: $owner, name: $repo) {
		milestones(query: $title, first: 30) {
			nodes {
				number
				id
				closed
				title
				closedAt
				dueOn
			}
		}
	}
}
`

func getMilestonesWithTitle(
	ctx context.Context,
	client graphql.Client,
	owner string,
	repo string,
	title string,
) (*getMilestonesWithTitleResponse, error) {
	req := &graphql.Request{
		OpName: "getMilestonesWithTitle",
		Query:  getMilestonesWithTitle_Operation,
		Variables: &__getMilestonesWithTitleInput{
			Owner: owner,
			Repo:  repo,
			Title: title,
		},
	}
	var err error

	var data getMilestonesWithTitleResponse
	resp := &graphql.Response{Data: &data}

	err = client.MakeRequest(
		ctx,
		req,
		resp,
	)

	return &data, err
}
