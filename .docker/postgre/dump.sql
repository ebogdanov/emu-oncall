SET client_encoding = 'UTF8';

ALTER DATABASE emu-oncall OWNER TO "user";

CREATE TYPE public.role_type AS ENUM (
    'user',
    'observer',
    'admin'
    );

ALTER TYPE public.role_type OWNER TO "user";

CREATE TABLE public.users (
  id integer NOT NULL,
  user_id character varying(15) NOT NULL,
  name character varying(100),
  username character varying(32) NOT NULL,
  email character varying(120),
  phone_number character varying(20),
  active boolean DEFAULT false,
  role public.role_type DEFAULT 'user'::public.role_type NOT NULL
);


ALTER TABLE public.users OWNER TO "user";

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO "user";

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

SELECT pg_catalog.setval('public.users_id_seq', 1, false);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT id_primary_key PRIMARY KEY (id);

CREATE INDEX email_index ON public.users USING btree (email);

CREATE INDEX index_role ON public.users USING btree (role);

CREATE INDEX index_user_id ON public.users USING btree (user_id);

create sequence public.events_id_seq;

alter sequence public.events_id_seq owner to "user";

alter sequence public.events_id_seq owned by events.id;


create table public.events (
   id bigint primary key not null default nextval('events_id_seq'::regclass),
   date_add timestamp without time zone,
   user_id character varying(50) not null,
   channel character varying(20) not null,
   recipient character varying(50) not null,
   success boolean not null default false,
   msg character varying(500)
);
create index index_date_add on events using btree (date_add);
create index index_success on events using btree (success);
create index index_recipient on events using btree (recipient);

create table public.client_devices (
   phone character varying(20) primary key not null,
   token character varying(512),
   date_activate date
);

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;

INSERT INTO public.users (id, user_id, name, username, email, phone_number, active, role) VALUES (1, 'SUPERTEST1', 'Evgeniy Bogdanov', 'e.bogdanov', 'e.bogdanov@biz-systems.ru', '+79281972661', true, 'admin');

create table client_devices
(
    phone         varchar(20) not null
        constraint primary_index
            primary key,
    token         varchar(512),
    date_activate date
);

alter table client_devices
    owner to iron_maiden;

