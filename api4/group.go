// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/model"
)

const (
	apiGroupMemberActionCreate = iota
	apiGroupMemberActionDelete
)

func (api *API) InitGroup() {
	api.BaseRoutes.Groups.Handle("", api.ApiSessionRequired(createGroup)).Methods("POST")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}", api.ApiSessionRequiredTrustRequester(getGroup)).Methods("GET")
	api.BaseRoutes.Groups.Handle("", api.ApiSessionRequired(getGroups)).Methods("GET")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}", api.ApiSessionRequired(updateGroup)).Methods("PUT")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}", api.ApiSessionRequired(deleteGroup)).Methods("DELETE")

	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/members", api.ApiSessionRequired(createOrDeleteGroupMember(apiGroupMemberActionCreate))).Methods("POST")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/members/{user_id:[A-Za-z0-9]+}", api.ApiSessionRequired(createOrDeleteGroupMember(apiGroupMemberActionDelete))).Methods("DELETE")

	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/teams", api.ApiSessionRequired(createGroupSyncable(model.GSTeam))).Methods("POST")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}", api.ApiSessionRequired(getGroupSyncable(model.GSTeam))).Methods("GET")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/teams", api.ApiSessionRequired(getGroupSyncables(model.GSTeam))).Methods("GET")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}", api.ApiSessionRequired(updateGroupSyncable(model.GSTeam))).Methods("PUT")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}", api.ApiSessionRequired(deleteGroupSyncable(model.GSTeam))).Methods("DELETE")

	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/channels", api.ApiSessionRequired(createGroupSyncable(model.GSChannel))).Methods("POST")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/channels/{channel_id:[A-Za-z0-9]+}", api.ApiSessionRequired(getGroupSyncable(model.GSChannel))).Methods("GET")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/channels", api.ApiSessionRequired(getGroupSyncables(model.GSChannel))).Methods("GET")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/channels/{channel_id:[A-Za-z0-9]+}", api.ApiSessionRequired(updateGroupSyncable(model.GSChannel))).Methods("PUT")
	api.BaseRoutes.Groups.Handle("/{group_id:[A-Za-z0-9]+}/channels/{channel_id:[A-Za-z0-9]+}", api.ApiSessionRequired(deleteGroupSyncable(model.GSChannel))).Methods("DELETE")
}

func createGroup(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireGroupId()
	if c.Err != nil {
		return
	}

	group := model.GroupFromJson(r.Body)
	if group == nil {
		c.SetInvalidParam("group")
		return
	}

	if c.App.License() == nil || !*c.App.License().Features.LDAP {
		c.Err = model.NewAppError("Api4.createGroup", "api.group.license.error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	group, err := c.App.CreateGroup(group)
	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)

	b, marshalErr := json.Marshal(group)
	if marshalErr != nil {
		c.Err = model.NewAppError("Api4.createGroup", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
		return
	}

	w.Write(b)
}

func getGroup(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireGroupId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	group, err := c.App.GetGroup(c.Params.GroupId)
	if err != nil {
		c.Err = err
		return
	}

	b, marshalErr := json.Marshal(group)
	if marshalErr != nil {
		c.Err = model.NewAppError("Api4.getGroup", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
		return
	}

	w.Write(b)
}

func getGroups(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	groups, err := c.App.GetGroupsPage(c.Params.Page, c.Params.PerPage)
	if err != nil {
		c.Err = err
		return
	}

	b, marshalErr := json.Marshal(groups)
	if marshalErr != nil {
		c.Err = model.NewAppError("Api4.getGroups", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
		return
	}

	w.Write(b)
}

func updateGroup(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireGroupId()
	if c.Err != nil {
		return
	}

	update := model.GroupFromJson(r.Body)
	if update == nil {
		c.SetInvalidParam("group")
		return
	}

	if c.App.License() == nil || !*c.App.License().Features.LDAP {
		c.Err = model.NewAppError("Api4.updateGroup", "api.group.license.error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	update.Id = c.Params.GroupId

	group, err := c.App.UpdateGroup(update)
	if err != nil {
		c.Err = err
		return
	}

	b, marshalErr := json.Marshal(group)
	if marshalErr != nil {
		c.Err = model.NewAppError("Api4.updateGroup", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
		return
	}

	w.Write(b)
}

func deleteGroup(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireGroupId()
	if c.Err != nil {
		return
	}

	if c.App.License() == nil || !*c.App.License().Features.LDAP {
		c.Err = model.NewAppError("Api4.deleteGroup", "api.group.license.error", nil, "", http.StatusNotImplemented)
		return
	}

	if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
		c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
		return
	}

	if _, err := c.App.DeleteGroup(c.Params.GroupId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func createOrDeleteGroupMember(action int) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		c.RequireUserId()
		if c.Err != nil {
			return
		}

		if c.App.License() == nil || !*c.App.License().Features.LDAP {
			c.Err = model.NewAppError("Api4.createOrDeleteGroupMember", "api.group.license.error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		var createOrDeleteF func(string, string) (*model.GroupMember, *model.AppError)
		var successStatus int
		switch action {
		case apiGroupMemberActionCreate:
			createOrDeleteF = c.App.CreateGroupMember
			successStatus = http.StatusCreated
		case apiGroupMemberActionDelete:
			createOrDeleteF = c.App.DeleteGroupMember
			successStatus = http.StatusOK
		default:
			return
		}

		groupMember, err := createOrDeleteF(c.Params.GroupId, c.Params.UserId)
		if err != nil {
			c.Err = err
			return
		}

		w.WriteHeader(successStatus)

		b, marshalErr := json.Marshal(groupMember)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.createOrDeleteGroupMember", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}

func createGroupSyncable(syncableType model.GroupSyncableType) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Err = model.NewAppError("Api4.createGroupSyncable", "api.group.io_error", nil, err.Error(), http.StatusNotImplemented)
		}

		var groupSyncable *model.GroupSyncable
		err = json.Unmarshal(body, &groupSyncable)
		if err != nil || groupSyncable == nil || groupSyncable.Type != syncableType {
			c.SetInvalidParam(fmt.Sprintf("Group[%s]", strings.Join(model.GroupSyncableTypes, "|")))
			return
		}

		if c.App.License() == nil || !*c.App.License().Features.LDAP {
			c.Err = model.NewAppError("Api4.createGroupSyncable", "api.group.license.error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		var appErr *model.AppError
		groupSyncable, appErr = c.App.CreateGroupSyncable(groupSyncable)
		if appErr != nil {
			c.Err = appErr
			return
		}

		w.WriteHeader(http.StatusCreated)

		var marshalErr error
		b, marshalErr := json.Marshal(groupSyncable)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.createGroupSyncable", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}

func getGroupSyncable(syncableType model.GroupSyncableType) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		var syncableID string
		switch syncableType {
		case model.GSTeam:
			c.RequireTeamId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.TeamId
		case model.GSChannel:
			c.RequireChannelId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.ChannelId
		default:
			c.SetInvalidParam(fmt.Sprintf("[%s]_id", strings.Join(model.GroupSyncableTypes, "|")))
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		groupSyncable, err := c.App.GetGroupSyncable(c.Params.GroupId, syncableID, syncableType)
		if err != nil {
			c.Err = err
			return
		}

		b, marshalErr := json.Marshal(groupSyncable)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.getGroupSyncable", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}

func getGroupSyncables(syncableType model.GroupSyncableType) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		groupSyncables, err := c.App.GetGroupSyncablesPage(c.Params.GroupId, syncableType, c.Params.Page, c.Params.PerPage)
		if err != nil {
			c.Err = err
			return
		}

		b, marshalErr := json.Marshal(groupSyncables)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.getGroupSyncables", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}

func updateGroupSyncable(syncableType model.GroupSyncableType) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		var syncableID string
		switch syncableType {
		case model.GSTeam:
			c.RequireTeamId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.TeamId
		case model.GSChannel:
			c.RequireChannelId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.ChannelId
		default:
			c.SetInvalidParam(fmt.Sprintf("[%s]_id", strings.Join(model.GroupSyncableTypes, "|")))
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Err = model.NewAppError("Api4.updateGroupSyncable", "api.group.io_error", nil, err.Error(), http.StatusNotImplemented)
		}

		var groupSyncablePatch *model.GroupSyncablePatch
		err = json.Unmarshal(body, &groupSyncablePatch)

		if err != nil || groupSyncablePatch == nil {
			c.SetInvalidParam(fmt.Sprintf("Group[%s]Patch", strings.Join(model.GroupSyncableTypes, "|")))
			return
		}

		if c.App.License() == nil || !*c.App.License().Features.LDAP {
			c.Err = model.NewAppError("Api4.updateGroupSyncable", "api.group.license.error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		var appErr *model.AppError
		var groupSyncable *model.GroupSyncable
		groupSyncable, appErr = c.App.UpdateGroupSyncable(&model.GroupSyncable{
			GroupId:    c.Params.GroupId,
			SyncableId: syncableID,
			AutoAdd:    groupSyncablePatch.AutoAdd,
			CanLeave:   groupSyncablePatch.CanLeave,
			Type:       syncableType,
		})
		if appErr != nil {
			c.Err = appErr
			return
		}

		b, marshalErr := json.Marshal(groupSyncable)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.updateGroupSyncable", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}

func deleteGroupSyncable(syncableType model.GroupSyncableType) func(*Context, http.ResponseWriter, *http.Request) {
	return func(c *Context, w http.ResponseWriter, r *http.Request) {
		c.RequireGroupId()
		if c.Err != nil {
			return
		}

		var syncableID string
		switch syncableType {
		case model.GSTeam:
			c.RequireTeamId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.TeamId
		case model.GSChannel:
			c.RequireChannelId()
			if c.Err != nil {
				return
			}
			syncableID = c.Params.ChannelId
		default:
			c.SetInvalidParam(fmt.Sprintf("[%s]_id", strings.Join(model.GroupSyncableTypes, "|")))
		}

		if c.App.License() == nil || !*c.App.License().Features.LDAP {
			c.Err = model.NewAppError("Api4.deleteGroupSyncable", "api.group.license.error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionTo(c.Session, model.PERMISSION_MANAGE_SYSTEM) {
			c.SetPermissionError(model.PERMISSION_MANAGE_SYSTEM)
			return
		}

		groupSyncable, err := c.App.DeleteGroupSyncable(c.Params.GroupId, syncableID, syncableType)
		if err != nil {
			c.Err = err
			return
		}

		b, marshalErr := json.Marshal(groupSyncable)
		if marshalErr != nil {
			c.Err = model.NewAppError("Api4.deleteGroupSyncable", "api.group.marshal_error", nil, marshalErr.Error(), http.StatusNotImplemented)
			return
		}

		w.Write(b)
	}
}
