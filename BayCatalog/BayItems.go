package BayCatalog

type BuyRequest struct {
	Item     string
	Cost     int64
	Response chan bool
}

const (
	ItemPickaxe     = "pickaxe"
	ItemVentilation = "ventilation"
	ItemCart        = "cart"
)

var Equipments = map[string]int64{
	"pickaxe":     3000,
	"ventilation": 15000,
	"cart":        50000,
}
