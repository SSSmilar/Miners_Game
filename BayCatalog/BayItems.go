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
