CREATE TABLE promotions
(
    uuid   UUID,
    price  Float,
    expire DateTime('UTC')
) ENGINE=MergeTree
ORDER BY uuid;