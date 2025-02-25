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

--name: get-token-symbol
-- $1: contract_address
SELECT token_symbol FROM tokens WHERE contract_address = $1 AND removed = false
