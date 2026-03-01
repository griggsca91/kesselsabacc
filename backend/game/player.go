package game

type ShiftToken string

const (
	TokenFreeDraw        ShiftToken = "free_draw"
	TokenRefund          ShiftToken = "refund"
	TokenGeneralTariff   ShiftToken = "general_tariff"
	TokenMarkdown        ShiftToken = "markdown"
	TokenImmunity        ShiftToken = "immunity"
	TokenMajorFraud      ShiftToken = "major_fraud"
	TokenCookTheBooks    ShiftToken = "cook_the_books"
	TokenDirectTransaction ShiftToken = "direct_transaction"
	TokenPrimeSabacc     ShiftToken = "prime_sabacc"
)

var StartingTokens = []ShiftToken{
	TokenFreeDraw,
	TokenRefund,
	TokenGeneralTariff,
}

type Player struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Chips           int          `json:"chips"`
	Invested        int          `json:"invested"` // chips put in this round
	SandCard        *Card        `json:"sandCard"`
	BloodCard       *Card        `json:"bloodCard"`
	ShiftTokens     []ShiftToken `json:"shiftTokens"`
	UsedTokens      []ShiftToken `json:"usedTokens"`
	Stood           bool         `json:"stood"`    // stood this turn
	Eliminated      bool         `json:"eliminated"`
	IsHost          bool         `json:"isHost"`
	// Impostor dice rolls for reveal (set at reveal phase)
	SandDiceRoll    int          `json:"sandDiceRoll"`
	BloodDiceRoll   int          `json:"bloodDiceRoll"`
}

func NewPlayer(id, name string, chips int, isHost bool) *Player {
	tokens := make([]ShiftToken, len(StartingTokens))
	copy(tokens, StartingTokens)
	return &Player{
		ID:          id,
		Name:        name,
		Chips:       chips,
		ShiftTokens: tokens,
		UsedTokens:  []ShiftToken{},
		IsHost:      isHost,
	}
}

func (p *Player) HasToken(t ShiftToken) bool {
	for _, tok := range p.ShiftTokens {
		if tok == t {
			return true
		}
	}
	return false
}

func (p *Player) UseToken(t ShiftToken) bool {
	for i, tok := range p.ShiftTokens {
		if tok == t {
			p.ShiftTokens = append(p.ShiftTokens[:i], p.ShiftTokens[i+1:]...)
			p.UsedTokens = append(p.UsedTokens, t)
			return true
		}
	}
	return false
}

func (p *Player) Invest(amount int) bool {
	if p.Chips < amount {
		return false
	}
	p.Chips -= amount
	p.Invested += amount
	return true
}

func (p *Player) Refund(amount int) {
	p.Chips += amount
	p.Invested -= amount
	if p.Invested < 0 {
		p.Invested = 0
	}
}

func (p *Player) ResetRound() {
	p.Invested = 0
	p.Stood = false
	p.SandCard = nil
	p.BloodCard = nil
	p.SandDiceRoll = 0
	p.BloodDiceRoll = 0
}
