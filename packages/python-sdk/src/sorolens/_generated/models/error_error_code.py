from enum import StrEnum


class ErrorErrorCode(StrEnum):
    FORBIDDEN = "FORBIDDEN"
    INTERNAL = "INTERNAL"
    INVALID_INPUT = "INVALID_INPUT"
    NOT_FOUND = "NOT_FOUND"
    PAYLOAD_TOO_LARGE = "PAYLOAD_TOO_LARGE"
    RATE_LIMITED = "RATE_LIMITED"
    TIMEOUT = "TIMEOUT"
    UNAUTHORIZED = "UNAUTHORIZED"
    UNSUPPORTED_MEDIA_TYPE = "UNSUPPORTED_MEDIA_TYPE"

    def __str__(self) -> str:
        return str(self.value)
