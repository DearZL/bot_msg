package server

import "github.com/gin-gonic/gin"

type actionRoute struct {
	action  string
	handler func(*gin.Context)
}

// actionRoutes 维护 OneBot action 到 HTTP handler 的映射。
// 新增服务接口时，只需要在这里登记 action 和处理函数。
func (s *Server) actionRoutes() []actionRoute {
	return []actionRoute{
		{action: "send_private_msg", handler: s.sendPrivateMsg},
		{action: "send_group_msg", handler: s.sendGroupMsg},
		{action: "send_msg", handler: s.sendMsg},
		{action: "delete_msg", handler: s.deleteMsg},
		{action: "get_msg", handler: s.getMsg},
		{action: "get_login_info", handler: s.getLoginInfo},
		{action: "get_status", handler: s.getStatus},
		{action: "get_version_info", handler: s.getVersionInfo},
		{action: "get_friend_list", handler: s.getFriendList},
		{action: "get_group_list", handler: s.getGroupList},
		{action: "get_group_member_list", handler: s.getGroupMemberList},
	}
}
