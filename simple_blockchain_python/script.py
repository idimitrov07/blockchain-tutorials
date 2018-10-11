new_transactions = [{'amount': '30', 'sender':'alice', 'receiver':'bob'},
               	{'amount': '55', 'sender':'bob', 'receiver':'alice'}]

# import sha256
from hashlib import sha256

# sets the amount of leading zeros that must be found in the hash produced by the nonce
difficulty = 2
nonce = 0


# creating the proof 
proof = sha256((str(nonce) + str(new_transactions)).encode()).hexdigest()

# printing proof
print(proof)

  
# finding a proof that has 2 leading zeros
while True:
  nonce = nonce + 1
  proof = sha256((str(nonce) + str(new_transactions)).encode()).hexdigest()
  if proof[:5] == "00000":
    break
  print(proof)
 
final_proof = proof
print(final_proof)

from blockchain import Blockchain

block_one_transactions = {"sender":"Alice", "receiver": "Bob", "amount":"50"}
block_two_transactions = {"sender": "Bob", "receiver":"Cole", "amount":"25"}
block_three_transactions = {"sender":"Alice", "receiver":"Cole", "amount":"35"}
fake_transactions = {"sender": "Bob", "receiver":"Cole, Alice", "amount":"25"}

local_blockchain = Blockchain()
local_blockchain.print_blocks()
local_blockchain.add_block(block_one_transactions)
local_blockchain.add_block(block_two_transactions)
local_blockchain.add_block(block_three_transactions)
local_blockchain.print_blocks()
local_blockchain.chain[2].transactions = fake_transactions
local_blockchain.validate_chain()