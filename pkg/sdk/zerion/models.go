package zerion

type (
	FungibleImplementation struct {
		ChainID string `json:"chain_id"`
		Address string `json:"address"`
	}

	FungibleInfo struct {
		Name            string                   `json:"name"`
		Symbol          string                   `json:"symbol"`
		Implementations []FungibleImplementation `json:"implementations"`
	}

	Attributes struct {
		Name         string       `json:"name"`
		PositionType string       `json:"position_type"`
		Value        float64      `json:"value"`
		Price        float64      `json:"price"`
		FungibleInfo FungibleInfo `json:"fungible_info"`
	}

	WalletPositionsData struct {
		Type       string     `json:"type"`
		ID         string     `json:"id"`
		Attributes Attributes `json:"attributes"`
	}

	WalletPositions struct {
		Positions []WalletPositionsData `json:"data"`
	}
)
