package omni

const (
	MP_TX_NOT_FOUND            = -3331 // No information available about transaction. (GetTransaction failed)
	MP_TX_IS_NOT_OMNI_PROTOCOL = -3336 // No Omni Layer Protocol transaction.
)

func OmniErrorInfo(errCode int) string {
	switch errCode {
	case MP_TX_NOT_FOUND:
		return "No information available about transaction"
	case MP_TX_IS_NOT_OMNI_PROTOCOL:
		return "No Omni Layer Protocol transaction"

	default:
		return "* unknown type *"
	}
}
