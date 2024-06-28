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

--name: insert-token-mint
-- $1: tx_id
-- $2: minter_address
-- $3: recipient_address
-- $4: mint_value
-- $5: contract_address
INSERT INTO token_mint(
    tx_id,
    minter_address,
    recipient_address,
    mint_value,
    contract_address
) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

--name: insert-token-burn
-- $1: tx_id
-- $2: burner_address
-- $3: burn_value
-- $4: contract_address
INSERT INTO token_burn(
    tx_id,
    burner_address,
    burn_value,
    contract_address
) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING

--name: insert-faucet-give
-- $1: tx_id
-- $2: token_address
-- $3: recipient_address
-- $4: give_value
-- $5: contract_address
INSERT INTO faucet_give(
    tx_id,
    token_address,
    recipient_address,
    give_value,
    contract_address
) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

--name: insert-pool-swap
-- $1: tx_id
-- $2: initiator_address
-- $3: token_in_address
-- $4: token_out_address
-- $5: in_value
-- $6: out_value
-- $7: fee
-- $8: contract_address
INSERT INTO pool_swap(
    tx_id,
    initiator_address,
    token_in_address,
    token_out_address,
    in_value,
    out_value,
    fee,
    contract_address
) VALUES($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING

--name: insert-pool-deposit
-- $1: tx_id
-- $2: initiator_address
-- $3: token_in_address
-- $4: in_value
-- $5: contract_address
INSERT INTO pool_deposit(
    tx_id,
    initiator_address,
    token_in_address,
    in_value,
    contract_address
) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING

--name: insert-price-quote-update
-- $1: tx_id
-- $2: token
-- $3: exchange_rate
-- $4: contract_address
INSERT INTO price_index_updates(
    tx_id,
    token,
    exchange_rate.
    contract_address
) VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING
