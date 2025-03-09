SET client_encoding = 'UTF8';

-- noinspection SqlResolve
ALTER DATABASE emu_oncall OWNER TO "user";

CREATE TYPE public.role_type AS ENUM (
    'user',
    'observer',
    'admin'
    );
ALTER TYPE public.role_type OWNER TO "user";

CREATE TABLE public.oncall_users
(
    id           integer                                           NOT NULL,
    user_id      character varying(15)                             NOT NULL,
    name         character varying(100),
    username     character varying(32)                             NOT NULL,
    email        character varying(120),
    phone_number character varying(20),
    active       boolean          DEFAULT false,
    role         public.role_type DEFAULT 'user'::public.role_type NOT NULL
);
ALTER TABLE public.oncall_users
    OWNER TO "user";

CREATE SEQUENCE public.oncall_users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.oncall_users_id_seq
    OWNER TO "user";

ALTER SEQUENCE public.oncall_users_id_seq OWNED BY public.oncall_users.id;

ALTER TABLE ONLY public.oncall_users
    ALTER COLUMN id SET DEFAULT nextval('public.oncall_users_id_seq'::regclass);

SELECT pg_catalog.setval('public.oncall_users_id_seq', 1, false);

ALTER TABLE ONLY public.oncall_users
    ADD CONSTRAINT id_primary_key PRIMARY KEY (id);

CREATE INDEX email_index ON public.oncall_users USING btree (email);
CREATE INDEX index_role ON public.oncall_users USING btree (role);
CREATE INDEX index_user_id ON public.oncall_users USING btree (user_id);

create sequence public.events_id_seq;

create table public.events
(
    id        bigint primary key    not null default nextval('events_id_seq'::regclass),
    date_add  timestamp without time zone,
    user_id   character varying(50) not null,
    channel   character varying(20) not null,
    recipient character varying(50) not null,
    success   boolean               not null default false,
    msg       character varying(500)
);

create index index_date_add on public.events using btree (date_add);
create index index_success on public.events using btree (success);
create index index_recipient on public.events using btree (recipient);

alter sequence public.events_id_seq owner to "user";
alter sequence public.events_id_seq owned by public.events.id;

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;
