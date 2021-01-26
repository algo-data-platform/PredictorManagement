package logics

const (
	TokenKey         = "vue_predictor_management_token"
	AdminRole        = "admin"
	AdminPassword    = "admin"
	AlgoUserRole     = "common_user"
	AlgoUserPassword = "common_user"
)

func IsValidUser(userName, password string) bool {
	switch userName {
	case AdminRole:
		if password == AdminPassword {
			return true
		}
	case AlgoUserRole:
		if password == AlgoUserPassword {
			return true
		}
	default:
		return false
	}
	return false
}
