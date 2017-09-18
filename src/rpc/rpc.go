package rpc

const (
    TRANSFER_DATA_MSG = iota
    FIND_CONTACT_MSG
    FIND_DATA_MSG
    STORE_DATA_MSG
    PING_MSG
    PONG_MSG
)

func EnumToString(enum int) string {
    switch enum {
    case TRANSFER_DATA_MSG:
        return "TRANSFER_DATA_MSG"
    case FIND_CONTACT_MSG:
        return "FIND_CONTACT_MSG"
    case FIND_DATA_MSG:
        return "FIND_DATA_MSG"
    case STORE_DATA_MSG:
        return "STORE_DATA_MSG"
    case PING_MSG:
        return "PING_MSG"
    case PONG_MSG:
        return "PONG_MSG"
    default:
        return "UNKNOWN_MSG"
    }
}
