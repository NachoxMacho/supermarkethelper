package types

type ProductItem struct {
	Name           string  `json:"name,omitempty"`
	Category       string  `json:"category,omitempty"`
	ID             int     `json:"id,omitempty"`
	ItemsPerBox    int     `json:"items_per_box,omitempty"`
	BoxPrice       float64 `json:"box_price,omitempty"`
	ItemsPerShelf  int     `json:"items_per_shelf,omitempty"`
	ShelvesInStore int     `json:"shelves_in_store,omitempty"`
}

type ProductItemOutput struct {
	Name           string `json:"name,omitempty"`
	Category       string `json:"category,omitempty"`
	ID             string `json:"id,omitempty"`
	ItemsPerBox    string `json:"items_per_box,omitempty"`
	BoxPrice       string `json:"box_price,omitempty"`
	ItemsPerShelf  string `json:"items_per_shelf,omitempty"`
	StockedAmount  string `json:"stocked_amount,omitempty"`
	ShelvesInStore string `json:"shelves_in_store,omitempty"`
	BoxesPerShelf  string `json:"boxes_per_shelf,omitempty"`
	PricePerItem   string `json:"price_per_item,omitempty"`
	SalePrice      string `json:"sale_price,omitempty"`
}
