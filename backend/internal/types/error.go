package types

import "errors"

var UserNotFoundError = errors.New("user not found")

const DatasourceNotFoundMsg = "Datasource not found"

const FailedToSetDatasourceActiveMsg = "Failed to set datasource active"

const InvalidUUIDMessage = "Invalid UUID"
