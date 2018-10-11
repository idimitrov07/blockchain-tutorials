from block import Block

class Blockchain:
  def __init__(self):
    self.chain = []
    self.all_transactions = []
    self.genesis_block()
  
  def genesis_block(self):
    transactions = {}
    gen_block = Block(transactions, '0')
    self.chain.append(gen_block)
    return self.chain
  
  # print blockchain
  def print_blocks(self):
    for i in range(len(self.chain)):
      current_block = self.chain[i]
      print("Block {} {}".format(i, current_block))
      current_block.print_block()
  
  # add block to the chain
  def add_block(self, transactions):
    previous_block_hash = self.chain[len(self.chain) - 1].generate_hash
    new_block = Block(transactions, previous_block_hash)
    self.chain.append(new_block)
  
  def validate_chain(self):
    for i in range(1, len(self.chain)):
      current = self.chain[i]
      previous = self.chain[i-1]
      if current.hash != current.generate_hash():
        return False
      if previous.hash != previous.generate_hash():
        return False
    return True

  # implement proof of work
  def proof_of_work(self,block, difficulty=2):
    diff_str = difficulty * "0"
    proof = block.generate_hash()
    while True:
      block.nonce = block.nonce + 1
      if proof[0:2] == diff_str:
        block.nonce = 0
        break
      else:
        proof = block.generate_hash()
    return proof
  
  # add new blocks
  def add_block(self, transactions):
    previous_block_hash = self.chain[len(self.chain) - 1].hash
    new_block = Block(transactions, previous_block_hash)
    proof = self.proof_of_work(new_block)
    self.chain.append(new_block)
    return(proof, new_block)






new_transactions = [{'amount': '30', 'sender':'alice', 'receiver':'bob'},
               	{'amount': '55', 'sender':'bob', 'receiver':'alice'}]

my_blockchain = Blockchain()
my_blockchain.add_block(new_transactions)
my_blockchain.print_blocks()

my_blockchain.chain[1].transactions = "fake_transactions" 
print("Chain is valid: {}".format(my_blockchain.validate_chain()))