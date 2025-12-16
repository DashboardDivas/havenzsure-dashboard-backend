package auth

type RoleCode string

var rolePermissions = map[RoleCode]map[string]struct{}{
	RoleSuperAdmin: {
		"workorders:list":    {},
		"workorders:getByID": {},
		"workorders:create":  {},
		//"workorders:update": {},

		//
	},
	RoleAdmin: {
		"workorders:list":    {},
		"workorders:getByID": {},
		"workorders:create":  {},
		//"workorders:update": {},
	},
	RoleAdjuster: {
		"workorders:list":    {},
		"workorders:getByID": {},
		"workorders:create":  {},
		// "workorders:update": {},
	},
	RoleBodyman: {
		"workorders:list":    {},
		"workorders:getByID": {},
		// "workorders:create": {},
		// "workorders:update": {},
	},
}

func Can(role RoleCode, permission string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	_, hasPerm := perms[permission]
	return hasPerm
}
