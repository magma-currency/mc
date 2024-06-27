# Transactioms in MoltenChain

Transactions in MoltenChain can be of two types:
a. Simple Transaction (Public input)
b. Zether Transaction (Confidential amount and Ring members)

Transaction Scripts in MoltenChain

a. Simple Transactions
  1. **SCRIPT_UPDATE_DELEGATE** will update delegate information and/or convert unclaimed funds into staking. 
  3. **SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY** will allow a liquidity offer for a certain asset. 
  
b. Zether Transaction
  1. **SCRIPT_TRANSFER** will transfer from an unknown sender to an unknown receiver an unknown amount. 
  4. **SCRIPT_ASSET_CREATE** will allow to create a new asset. The fee is paid by an unknown sender
  5. **SCRIPT_ASSET_SUPPLY_INCREASE** will allow to increase the supply of an asset X with value Y and move these to a known receiver address Z. The fee is paid by an unknown sender   
