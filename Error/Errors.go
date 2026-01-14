package Error

import "errors"

var ErrInvalidVariable = errors.New("Invalid_Variable")
var ErrNotEnoughMoney = errors.New("Not_Enough_Money")
var ErrAlreadyHaveItem = errors.New("Already_Have_Item")
var ErrWrongMethod = errors.New("Wrong_Method")
var ErrInvalidParameters = errors.New("invalid parameters")
var ErrBadRequest = errors.New("Bad_Request")
var ErrDataBase = errors.New("Data_Base_Error")
var ErrMinersDB = errors.New("Miners_not_found_or_alredy_delete")
var ErrTransaction = errors.New("Transaction_Error")
var ErrRegiments = errors.New("Regiments_Error")
var ErrInsert = errors.New("Insert_Error")
var ErrSaveChange = errors.New("Saved_Change_Error")
