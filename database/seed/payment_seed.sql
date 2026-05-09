INSERT INTO
    transactions (
        request_id,
        customer_pan,
        amount,
        transaction_datetime,
        rrn,
        bill_number,
        customer_name,
        merchant_id,
        merchant_name,
        merchant_city,
        currency_code,
        payment_status,
        payment_description
    )
VALUES (
        'XwVjF5zfuHhrDZuw',
        '9360001110000000019',
        10000.00,
        '2021-02-25T13:36:13Z',
        '123456789012',
        '12345678901234567890',
        'John Doe',
        '008800223497',
        'Sukses Makmur Bendungan Hilir',
        'Jakarta Pusat',
        '360',
        '00',
        'Payment Success'
    ) ON CONFLICT (request_id) DO NOTHING;