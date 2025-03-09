-- Sample of file. Here you can add additional records, create tables, etc
INSERT INTO public.oncall_users (user_id, name, username, email, phone_number, active, role)
VALUES ('UGRAFANAEMU1', 'Laura Preston', 'LauraPreston@protonmail.com', 'LauraPreston@protonmail.com', '+79000000000', true, 'admin'),
ON CONFLICT (id) DO NOTHING;
