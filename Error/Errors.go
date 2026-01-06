package Error

import "errors"

var ErrInvalidVariable = errors.New("Invalid_Variable")
var ErrNotEnoughMoney = errors.New("Not_Enough_Money")
var ErrAlreadyHaveItem = errors.New("Already_Have_Item")
