-- 0002_profile_seed.sql — Fila única del perfil (id fijo 1)
-- Garantiza que GET /api/v1/profile siempre tenga datos.

INSERT INTO profile (id, name)
VALUES (1, '')
ON CONFLICT (id) DO NOTHING;
