from datetime import datetime
from common.utils import Bet

def decode_bet(raw_message: str) -> Bet:
    """
    Decode a bet string into a Bet object.
    """
    bet_fields = raw_message.split("|")

    return Bet(
        bet_fields[0],
        bet_fields[1],
        last_name=bet_fields[2],
        document=bet_fields[3],
        birthdate=bet_fields[4],
        number=bet_fields[5]
    )

def decode_bet_batch(raw_message: str) -> list[Bet]:
    """
    Decode a batch of bets from a raw message string.
    """
    if raw_message.startswith("BET_BATCH:"):
        raw_message = raw_message[len("BET_BATCH:"):]
        
    bet_strings = raw_message.strip().split("\n")
    bets = []

    for encoded_bet in bet_strings:
        if encoded_bet:
            bets.append(decode_bet(encoded_bet))

    return bets