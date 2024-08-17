--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3 (Debian 16.3-1.pgdg120+1)
-- Dumped by pg_dump version 16.3 (Debian 16.3-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_§th', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: aircraft_data; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.aircraft_data (
    id integer NOT NULL,
    hex character varying,
    flight character varying,
    first_seen timestamp with time zone,
    first_seen_epoch bigint,
    last_seen timestamp with time zone,
    last_seen_epoch bigint,
    type character varying,
    r character varying,
    t character varying,
    alt_baro integer,
    alt_geom integer,
    gs numeric(6,1),
    ias integer,
    tas integer,
    mach numeric(5,3),
    track numeric(5,2),
    track_rate numeric(5,2),
    roll numeric(5,2),
    mag_heading numeric(5,2),
    true_heading numeric(5,2),
    baro_rate integer,
    geom_rate integer,
    squawk character varying,
    emergency character varying,
    nav_qnh numeric(7,1),
    nav_altitude_mcp integer,
    nav_heading numeric(5,2),
    nav_modes text[],
    lat numeric(9,6),
    lon numeric(9,6),
    nic integer,
    rc integer,
    seen_pos numeric(9,3),
    r_dst numeric(8,3),
    r_dir numeric(8,3),
    version integer,
    nic_baro integer,
    nac_p integer,
    nac_v integer,
    sil integer,
    sil_type character varying,
    gva integer,
    sda integer,
    alert integer,
    spi integer,
    mlat text[],
    tisb text[],
    messages integer,
    seen numeric(8,3),
    rssi numeric(6,1),
    highest_aircraft_processed boolean DEFAULT false,
    lowest_aircraft_processed boolean DEFAULT false,
    fastest_aircraft_processed boolean DEFAULT false,
    slowest_aircraft_processed boolean DEFAULT false
);


ALTER TABLE public.aircraft_data OWNER TO admin;

--
-- Name: aircraft_data_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.aircraft_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.aircraft_data_id_seq OWNER TO admin;

--
-- Name: aircraft_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.aircraft_data_id_seq OWNED BY public.aircraft_data.id;


--
-- Name: fastest_aircraft; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.fastest_aircraft (
    id integer NOT NULL,
    hex character varying,
    flight character varying,
    registration character varying,
    type character varying,
    first_seen timestamp with time zone,
    last_seen timestamp with time zone,
    ground_speed numeric(6,1),
    indicated_air_speed integer,
    true_air_speed integer
);


ALTER TABLE public.fastest_aircraft OWNER TO admin;

--
-- Name: fastest_aircraft_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.fastest_aircraft_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.fastest_aircraft_id_seq OWNER TO admin;

--
-- Name: fastest_aircraft_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.fastest_aircraft_id_seq OWNED BY public.fastest_aircraft.id;


--
-- Name: highest_aircraft; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.highest_aircraft (
    id integer NOT NULL,
    hex character varying,
    flight character varying,
    registration character varying,
    type character varying,
    first_seen timestamp with time zone,
    last_seen timestamp with time zone,
    barometric_altitude integer,
    geometric_altitude integer
);


ALTER TABLE public.highest_aircraft OWNER TO admin;

--
-- Name: highest_aircraft_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.highest_aircraft_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.highest_aircraft_id_seq OWNER TO admin;

--
-- Name: highest_aircraft_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.highest_aircraft_id_seq OWNED BY public.highest_aircraft.id;


--
-- Name: lowest_aircraft; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.lowest_aircraft (
    id integer NOT NULL,
    hex character varying,
    flight character varying,
    registration character varying,
    type character varying,
    first_seen timestamp with time zone,
    last_seen timestamp with time zone,
    barometric_altitude integer,
    geometric_altitude integer
);


ALTER TABLE public.lowest_aircraft OWNER TO admin;

--
-- Name: lowest_aircraft_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.lowest_aircraft_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.lowest_aircraft_id_seq OWNER TO admin;

--
-- Name: lowest_aircraft_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.lowest_aircraft_id_seq OWNED BY public.lowest_aircraft.id;


--
-- Name: slowest_aircraft; Type: TABLE; Schema: public; Owner: admin
--

CREATE TABLE public.slowest_aircraft (
    id integer NOT NULL,
    hex character varying,
    flight character varying,
    registration character varying,
    type character varying,
    first_seen timestamp with time zone,
    last_seen timestamp with time zone,
    ground_speed numeric(6,1),
    indicated_air_speed integer,
    true_air_speed integer
);


ALTER TABLE public.slowest_aircraft OWNER TO admin;

--
-- Name: slowest_aircraft_id_seq; Type: SEQUENCE; Schema: public; Owner: admin
--

CREATE SEQUENCE public.slowest_aircraft_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.slowest_aircraft_id_seq OWNER TO admin;

--
-- Name: slowest_aircraft_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: admin
--

ALTER SEQUENCE public.slowest_aircraft_id_seq OWNED BY public.slowest_aircraft.id;


--
-- Name: test; Type: VIEW; Schema: public; Owner: admin
--

CREATE VIEW public.test AS
 SELECT id,
    r,
    hex,
    flight,
    first_seen,
    last_seen,
    alt_baro
   FROM public.aircraft_data
  ORDER BY last_seen DESC;


ALTER VIEW public.test OWNER TO admin;

--
-- Name: aircraft_data id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.aircraft_data ALTER COLUMN id SET DEFAULT nextval('public.aircraft_data_id_seq'::regclass);


--
-- Name: fastest_aircraft id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.fastest_aircraft ALTER COLUMN id SET DEFAULT nextval('public.fastest_aircraft_id_seq'::regclass);


--
-- Name: highest_aircraft id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.highest_aircraft ALTER COLUMN id SET DEFAULT nextval('public.highest_aircraft_id_seq'::regclass);


--
-- Name: lowest_aircraft id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.lowest_aircraft ALTER COLUMN id SET DEFAULT nextval('public.lowest_aircraft_id_seq'::regclass);


--
-- Name: slowest_aircraft id; Type: DEFAULT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.slowest_aircraft ALTER COLUMN id SET DEFAULT nextval('public.slowest_aircraft_id_seq'::regclass);


--
-- Name: fastest_aircraft fastest_aircraft_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.fastest_aircraft
    ADD CONSTRAINT fastest_aircraft_pkey PRIMARY KEY (id);


--
-- Name: fastest_aircraft fastest_aircraft_unique_hex_first_seen; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.fastest_aircraft
    ADD CONSTRAINT fastest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen);


--
-- Name: highest_aircraft highest_aircraft_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.highest_aircraft
    ADD CONSTRAINT highest_aircraft_pkey PRIMARY KEY (id);


--
-- Name: highest_aircraft highest_aircraft_unique_hex_first_seen; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.highest_aircraft
    ADD CONSTRAINT highest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen);


--
-- Name: lowest_aircraft lowest_aircraft_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.lowest_aircraft
    ADD CONSTRAINT lowest_aircraft_pkey PRIMARY KEY (id);


--
-- Name: lowest_aircraft lowest_aircraft_unique_hex_first_seen; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.lowest_aircraft
    ADD CONSTRAINT lowest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen);


--
-- Name: slowest_aircraft slowest_aircraft_pkey; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.slowest_aircraft
    ADD CONSTRAINT slowest_aircraft_pkey PRIMARY KEY (id);


--
-- Name: slowest_aircraft slowest_aircraft_unique_hex_first_seen; Type: CONSTRAINT; Schema: public; Owner: admin
--

ALTER TABLE ONLY public.slowest_aircraft
    ADD CONSTRAINT slowest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen);


--
-- PostgreSQL database dump complete
--