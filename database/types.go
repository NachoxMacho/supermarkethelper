package database

type Product struct {
	Name          string
	ID            int
	CategoryID    int
	ItemsPerBox   int
	ItemsPerShelf int
	DefaultBoxPrice float64
	DefaultShelvesInStore int
}

type Category struct {
	Name      string
	ID        int
	Precursor int
}

type Session struct {
	ID string
}

type SessionCategory struct {
	SessionID string
	CategoryID int
}

type SessionProductSpecific struct {
	SessionID      string
	ProductID      int
	BoxPrice       float64
	ShelvesInStore int
}
