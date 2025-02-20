--name: insert-tx
-- $1: tx_hash
-- $2: block_number
-- $3: date_block
-- $4: success
WITH insert_tx AS (
    INSERT INTO tx(
        tx_hash,
        block_number,
        date_block,
        success
    ) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id
)
SELECT id FROM insert_tx
UNION ALL
SELECT id FROM tx WHERE tx_hash = $1 AND id IS NOT NULL
LIMIT 1

--name: insert-token-transfer
-- $1: tx_id
-- $2: sender_address
-- $3: recipient_address
-- $4: transfer_value
-- $5: contract_address
INSERT INTO token_transfer(
    tx_id,
    sender_address,
    recipient_address,
    transfer_value,
    contract_address
) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

-- --name: insert-token-mint
-- -- $1: tx_id
-- -- $2: minter_address
-- -- $3: recipient_address
-- -- $4: mint_value
-- -- $5: contract_address
-- INSERT INTO token_mint(
--     tx_id,
--     minter_address,
--     recipient_address,
--     mint_value,
--     contract_address
-- ) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

-- --name: insert-token-burn
-- -- $1: tx_id
-- -- $2: burner_address
-- -- $3: burn_value
-- -- $4: contract_address
-- INSERT INTO token_burn(
--     tx_id,
--     burner_address,
--     burn_value,
--     contract_address
-- ) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING

-- --name: insert-faucet-give
-- -- $1: tx_id
-- -- $2: token_address
-- -- $3: recipient_address
-- -- $4: give_value
-- -- $5: contract_address
-- INSERT INTO faucet_give(
--     tx_id,
--     token_address,
--     recipient_address,
--     give_value,
--     contract_address
-- ) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

-- --name: insert-pool-swap
-- -- $1: tx_id
-- -- $2: initiator_address
-- -- $3: token_in_address
-- -- $4: token_out_address
-- -- $5: in_value
-- -- $6: out_value
-- -- $7: fee
-- -- $8: contract_address
-- INSERT INTO pool_swap(
--     tx_id,
--     initiator_address,
--     token_in_address,
--     token_out_address,
--     in_value,
--     out_value,
--     fee,
--     contract_address
-- ) VALUES($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING

-- --name: insert-pool-deposit
-- -- $1: tx_id
-- -- $2: initiator_address
-- -- $3: token_in_address
-- -- $4: in_value
-- -- $5: contract_address
-- INSERT INTO pool_deposit(
--     tx_id,
--     initiator_address,
--     token_in_address,
--     in_value,
--     contract_address
-- ) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

-- --name: insert-ownership-change
-- -- $1: tx_id
-- -- $2: previous_owner
-- -- $3: new_owner
-- -- $4: contract_address
-- INSERT INTO ownership_change(
--     tx_id,
--     previous_owner,
--     new_owner,
--     contract_address
-- ) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING

--name: insert-token
-- $1: contract_address
-- $2: token_name
-- $3: token_symbol
-- $4: token_decimals
-- $5: sink_address
INSERT INTO tokens(
	contract_address,
	token_name,
	token_symbol,
	token_decimals,
    sink_address
) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

-- --name: insert-pool
-- -- $1: contract_address
-- -- $2: pool_name
-- -- $3: pool_symbol
-- INSERT INTO pools(
-- 	contract_address,
--     pool_name,
--     pool_symbol
-- ) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING

-- --name: remove-pool
-- -- $1: contract_address
-- UPDATE pools SET removed = true WHERE contract_address = $1

-- --name: remove-token
-- -- $1: contract_address
-- UPDATE tokens SET removed = true WHERE contract_address = $1
