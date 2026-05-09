INSERT INTO
    merchants (
        merchant_id,
        name,
        secret_hash,
        signing_key,
        is_active
    )
VALUES (
        '008800223497',
        'Sukses Makmur Bendungan Hilir',
        '$2a$10$xNR6OnCT2/W08JWmvIJoaOusIaROPbHQwoLkoqeUdCfQVLrcRk2B6',
        'speskilltest',
        TRUE
    ) ON CONFLICT (merchant_id) DO NOTHING;