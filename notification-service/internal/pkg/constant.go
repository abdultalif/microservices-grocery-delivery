package pkg

import "strconv"

const (
	NOTIF_EMAIL_VERIFICATION        = "email_verification"
	NOTIF_EMAIL_FORGOT_PASSWORD     = "reset_password"
	NOTIF_EMAIL_CREATE_CUSTOMER     = "create_customer"
	NOTIF_EMAIL_UPDATE_CUSTOMER     = "update_customer"
	NOTIF_EMAIL_UPDATE_STATUS_ORDER = "email-update-status-order"
	PUSH_NOTIF                      = "push-notif"
)

func StringToInt64(s string) (int64, error) {
	newData, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return newData, nil
}
