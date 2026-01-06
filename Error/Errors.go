package Error

import "errors"

var ErrInvalidVariable = errors.New("Invalid_Variable")
var ErrNotEnoughMoney = errors.New("Not_Enough_Money")
var ErrAlreadyHaveItem = errors.New("Already_Have_Item")
var ErrWrongMethod = errors.New("Wrong_Method")
var ErrInvalidParameters = errors.New("invalid parameters")
var ErrBadRequest = errors.New("Bad_Request")
