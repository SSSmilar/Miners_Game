package BayCatalog

type BuyRequest struct {
	Item     string
	Cost     int64
	Response chan bool
}
type HireRequest struct {
	MinerType string
	Cost      int64
	Response  chan bool
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

var WorkForce = map[string]int64{
	"tiny":   50,
	"medium": 200,
	"strong": 1000,
}
