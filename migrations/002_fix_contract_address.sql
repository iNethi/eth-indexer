ALTER TABLE tx DROP COLUMN contract_address;

ALTER TABLE token_transfer ADD COLUMN contract_address VARCHAR(42);
UPDATE token_transfer SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE token_transfer ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE token_mint ADD COLUMN contract_address VARCHAR(42);
UPDATE token_mint SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE token_mint ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE token_burn ADD COLUMN contract_address VARCHAR(42);
UPDATE token_burn SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE token_burn ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE faucet_give ADD COLUMN contract_address VARCHAR(42);
UPDATE faucet_give SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE faucet_give ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE pool_swap ADD COLUMN contract_address VARCHAR(42);
UPDATE pool_swap SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE pool_swap ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE pool_deposit ADD COLUMN contract_address VARCHAR(42);
UPDATE pool_deposit SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE pool_deposit ALTER COLUMN contract_address SET NOT NULL;

ALTER TABLE price_index_updates ADD COLUMN contract_address VARCHAR(42);
UPDATE price_index_updates SET contract_address = '0x0000000000000000000000000000000000000000';
ALTER TABLE price_index_updates ALTER COLUMN contract_address SET NOT NULL;