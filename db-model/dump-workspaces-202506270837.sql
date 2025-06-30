--
-- PostgreSQL database dump
--

-- Dumped from database version 14.2
-- Dumped by pg_dump version 15.3

-- Started on 2025-06-27 08:37:51

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE workspaces;
--
-- TOC entry 5577 (class 1262 OID 16385)
-- Name: workspaces; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE workspaces WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'C';


ALTER DATABASE workspaces OWNER TO postgres;

\connect workspaces

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 12 (class 2615 OID 38826)
-- Name: workspace_s1v8h2yrq9x15u1x; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA workspace_s1v8h2yrq9x15u1x;


ALTER SCHEMA workspace_s1v8h2yrq9x15u1x OWNER TO postgres;

--
-- TOC entry 934 (class 1255 OID 39938)
-- Name: compute_id_from_parent_and_name(uuid, text, text); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.compute_id_from_parent_and_name(parent uuid, table_name text, item_name text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    id uuid;
BEGIN
    id = md5(parent || '|' || table_name || '|' || item_name)::uuid;
    RETURN id;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.compute_id_from_parent_and_name(parent uuid, table_name text, item_name text) OWNER TO postgres;

--
-- TOC entry 947 (class 1255 OID 40083)
-- Name: copyto_aliases_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_aliases_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
return format('COPY (select a.* from aliases a
        where a.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_aliases_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 890 (class 1255 OID 39060)
-- Name: copyto_blocks_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_blocks_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN 'COPY (
               SELECT b.* FROM blocks b WHERE b.id IN (SELECT id FROM block_ids_with_refs)
           ) TO STDOUT (FORMAT CSV, HEADER)';
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_blocks_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 891 (class 1255 OID 39061)
-- Name: copyto_boards_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_boards_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select b.* from boards b
        left join projects p on b.project_id = p.id
        where p.id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_boards_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 930 (class 1255 OID 39839)
-- Name: copyto_connections_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_connections_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select distinct c.* from connections c
        where c.project_id = %L
    ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_connections_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 929 (class 1255 OID 39833)
-- Name: copyto_files_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_files_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN format('COPY (
                SELECT f.id, f.project_id, NULL::bytea as content, f.name, f.mime_type, f.size, f.created_at, f.modified_at FROM files f
                    WHERE f.id IN (SELECT UNNEST(list_related_files(%L)))
           ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_files_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 931 (class 1255 OID 39862)
-- Name: copyto_io_assignments_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_io_assignments_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN FORMAT('COPY (select i.* from io_assignments i
        left join levels l on i.model_level_id = l.id
        left join models m on l.model_id = m.id
        where m.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_io_assignments_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 892 (class 1255 OID 39063)
-- Name: copyto_levels_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_levels_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN format('COPY (SELECT l.* FROM levels l
                          LEFT JOIN models m ON l.model_id = m.id
                          WHERE m.project_id = %L
                       UNION
                        (WITH RECURSIVE cte(id, parent_id) AS (
                          SELECT l.id, l.parent_id FROM levels l
                            INNER JOIN (SELECT DISTINCT l.id FROM levels l
                              JOIN blocks b ON b.level_id = l.id OR (b.level_id IS NULL AND b.subsystem_level_id = l.id)
                              JOIN block_ids_with_refs bir ON bir.id = b.id
                              JOIN models m ON m.id = l.model_id
                              WHERE m.project_id != %L
                            ) AS x(id) ON l.id = x.id
                        UNION ALL
                          SELECT l.id, l.parent_id FROM levels l
                            INNER JOIN cte ON cte.parent_id = l.id
                        ) SELECT l.* FROM levels l
                            JOIN cte on l.id = cte.id
                      )
                   ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id, lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_levels_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 893 (class 1255 OID 39064)
-- Name: copyto_links_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_links_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select links.* from links
        left join levels l on links.level_id = l.id
        left join models m on l.model_id = m.id
        where m.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_links_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 894 (class 1255 OID 39065)
-- Name: copyto_migrations_query(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_migrations_query() RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return 'COPY (select * from migrations) TO STDOUT (FORMAT CSV, HEADER)';
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_migrations_query() OWNER TO postgres;

--
-- TOC entry 895 (class 1255 OID 39066)
-- Name: copyto_models_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_models_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN format('COPY (SELECT m.* FROM models m
                           WHERE m.project_id = %L
                       UNION
                         SELECT DISTINCT m.* FROM models m
                           JOIN levels l ON l.model_id = m.id
                           JOIN blocks b ON b.level_id = l.id OR (b.level_id IS NULL AND b.subsystem_level_id = l.id)
                           JOIN block_ids_with_refs bir on bir.id = b.id
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_models_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 928 (class 1255 OID 39752)
-- Name: copyto_nodes_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_nodes_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select n.* from nodes n
        left join models m on n.model_id = m.id
        where m.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_nodes_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 896 (class 1255 OID 39067)
-- Name: copyto_parameter_set_values_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameter_set_values_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select psv.* from parameter_set_values psv
        left join parameter_sets ps on psv.parameter_set_id = ps.id
        where ps.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameter_set_values_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 897 (class 1255 OID 39068)
-- Name: copyto_parameter_sets_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameter_sets_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select ps.* from parameter_sets ps
        where project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameter_sets_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 898 (class 1255 OID 39069)
-- Name: copyto_parameters_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameters_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN 'COPY (SELECT p.* FROM parameters p
                    JOIN block_ids_with_refs bir on bir.id = p.block_id
        ) TO STDOUT (FORMAT CSV, HEADER)';
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_parameters_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 899 (class 1255 OID 39070)
-- Name: copyto_ports_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_ports_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN 'COPY (SELECT p.* FROM ports p
                    JOIN block_ids_with_refs bir on bir.id = p.block_id
        ) TO STDOUT (FORMAT CSV, HEADER)';
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_ports_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 900 (class 1255 OID 39071)
-- Name: copyto_projects_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_projects_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN format('COPY (SELECT p.* FROM projects p
                           WHERE p.id = %L
                       UNION
                         SELECT DISTINCT p.* FROM projects p
                           JOIN models m on m.project_id = p.id
                           JOIN levels l ON l.model_id = m.id
                           JOIN blocks b ON b.level_id = l.id OR (b.level_id IS NULL AND b.subsystem_level_id = l.id)
                           JOIN block_ids_with_refs bir on bir.id = b.id
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_projects_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 936 (class 1255 OID 39961)
-- Name: copyto_scripts_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_scripts_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN FORMAT('COPY (select s.* from scripts s
        where s.project_id = %L) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_scripts_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 901 (class 1255 OID 39072)
-- Name: copyto_signals_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_signals_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    RETURN 'COPY (SELECT s.* FROM signals s
                    JOIN block_ids_with_refs bir on bir.id = s.block_id
        ) TO STDOUT (FORMAT CSV, HEADER)';
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_signals_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 911 (class 1255 OID 39453)
-- Name: copyto_simulation_configurations_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_simulation_configurations_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select s.* from simulation_configurations s
        where s.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_simulation_configurations_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 902 (class 1255 OID 39073)
-- Name: copyto_table_datapoints_query(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_table_datapoints_query(lookup_project_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
begin
    return format('COPY (select td.* from table_datapoints td
        left join boards b on td.board_id = b.id
        where b.project_id = %L
        ) TO STDOUT (FORMAT CSV, HEADER)', lookup_project_id);
end
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.copyto_table_datapoints_query(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 950 (class 1255 OID 40123)
-- Name: create_table_block_ids_with_refs(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.create_table_block_ids_with_refs(lookup_project_id uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    CREATE TEMP TABLE IF NOT EXISTS block_ids_with_refs AS (
        WITH RECURSIVE cte(id, ref_id) AS (
            SELECT b.id AS id, b.ref_id AS ref_id
            FROM blocks b
                JOIN levels l ON l.id = b.level_id OR (b.level_id IS NULL AND l.id = b.subsystem_level_id)
                JOIN models m ON m.id = l.model_id
            WHERE m.project_id = lookup_project_id
        UNION
            SELECT DISTINCT b.id AS id, b.ref_id AS ref_id
            FROM blocks b JOIN cte ON cte.ref_id = b.id
            WHERE cte.id != b.id
        )
        SELECT id FROM cte WHERE (id NOT IN (SELECT id FROM library_block_tree)) OR (lookup_project_id = '81a74576-35a9-47c1-b88d-97cd271f1c9f')
    );
    CREATE INDEX IF NOT EXISTS block_ids_with_refs_idx on block_ids_with_refs(id);
    ANALYZE block_ids_with_refs;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.create_table_block_ids_with_refs(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 937 (class 1255 OID 39964)
-- Name: create_table_block_ids_with_refs(uuid, boolean); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.create_table_block_ids_with_refs(lookup_project_id uuid, include_native_lib boolean) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    CREATE TEMP TABLE IF NOT EXISTS block_ids_with_refs AS (
        WITH RECURSIVE cte(id, ref_id) AS (
            SELECT b.id AS id, b.ref_id AS ref_id
            FROM blocks b
                JOIN levels l ON l.id = b.level_id OR (b.level_id IS NULL AND l.id = b.subsystem_level_id)
                JOIN models m ON m.id = l.model_id
            WHERE m.project_id = lookup_project_id
        UNION
            SELECT DISTINCT b.id AS id, b.ref_id AS ref_id
            FROM blocks b JOIN cte ON cte.ref_id = b.id
            WHERE cte.id != b.id
        )
        SELECT id FROM cte WHERE (id NOT IN (SELECT id FROM library_block_tree)) OR include_native_lib
    );
    CREATE INDEX IF NOT EXISTS block_ids_with_refs_idx on block_ids_with_refs(id);
    ANALYZE block_ids_with_refs;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.create_table_block_ids_with_refs(lookup_project_id uuid, include_native_lib boolean) OWNER TO postgres;

--
-- TOC entry 932 (class 1255 OID 39926)
-- Name: datapoints(uuid[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints(datapoint_ids uuid[]) RETURNS TABLE(id uuid, name text, type text, size integer[], data_type text, ref_id uuid, model_id uuid, model_name text, model_type text, block_id uuid, block_path text, block_name text, path text, alias text, default_value double precision[], default_value_text text, default_value_object jsonb, is_favorite boolean, is_visible boolean, is_tunable boolean, description text, is_read_only boolean, unit text, element_names jsonb, enum_values jsonb, folder_id uuid, pos integer, project_id uuid, file_id uuid, script_id uuid)
    LANGUAGE sql
    AS $$
WITH lwp (datapoint_id, id, path, model_id) AS (WITH datapoints (idx) AS (SELECT * FROM UNNEST(datapoint_ids) AS x(idx))
                                                SELECT datapoints.idx as datapoint_id, *
                                                FROM levels_with_path('/'::char, (SELECT array_agg(level_id)
                                                                                  FROM (SELECT b.level_id
                                                                                        FROM parameters p
                                                                                                 JOIN blocks b ON b.id = p.block_id
                                                                                                 JOIN datapoints ON p.id = datapoints.idx
                                                                                        UNION
                                                                                        SELECT b.level_id
                                                                                        FROM signals s
                                                                                                 JOIN blocks b ON b.id = s.block_id
                                                                                                 JOIN datapoints ON s.id = datapoints.idx) AS datapoint_levels)::UUID[]
                                                    ), datapoints
                                                WHERE path IS NOT NULL)
SELECT p.id                                                            AS id,
       p.name                                                          AS name,
       'PARAMETER'                                                     AS type,
       p.size                                                          AS size,
       p.data_type                                                     AS data_type,
       p.ref_id                                                        AS ref_id,
       m.id                                                            AS model_id,
       case
           when m.board_id is NULL then m.name
           else boards.name
           end                                                         AS model_name,
       m.type                                                          AS model_type,
       p.block_id,
       case
           when lwp.path like concat(m.name, '/%') then RIGHT(lwp.path, LENGTH(lwp.path) - LENGTH(m.name) -1)
           else lwp.path end                                           AS block_path,
       b.name                                                          AS block_name,
       CONCAT_WS('/', NULLIF(lwp.path, ''), b.name, p.name)            AS path,
       ''                                                              AS alias,
       p.default_value                                                 AS default_value,
       p.default_value_text                                            AS default_value_text,
       p.default_value_object                                          AS default_value_object,
       p.is_favorite                                                   AS is_favorite,
       p.is_visible                                                    AS is_visible,
       p.is_tunable                                                    AS is_tunable,
       case
           when p.ref_id is NULL then
               p.library_data::jsonb ->> 'description'
           else
               pref.library_data::jsonb ->> 'description'
           end
                                                                       AS description,
       p.is_read_only                                         AS is_read_only,
       case
           when p.ref_id is NULL then
               p.library_data::jsonb ->> 'unit'
           else
               pref.library_data::jsonb ->> 'unit'
           end
                                                              AS unit,
       case
           when p.ref_id is NULL then
               p.library_data::jsonb -> 'elementNames'
           else
               pref.library_data::jsonb -> 'elementNames'
           end
                                                              AS element_names,
       case
           when p.ref_id is NULL then
               p.library_data::jsonb -> 'enumValues'
           else
               pref.library_data::jsonb -> 'enumValues'
           end
                                                              AS enum_values,
       case
           when p.ref_id is NULL then
               (p.library_data::jsonb ->> 'folderId')::uuid
           else
               (pref.library_data::jsonb ->> 'folderId')::uuid
           end
                                                              AS folder_id,
       case
           when p.ref_id is NULL then
               (p.library_data::jsonb ->> 'position')::int
           else
               (pref.library_data::jsonb ->> 'position')::int
           end
                                                              AS pos,
       m.project_id                                           AS project_id,
       p.file_id                                              AS file_id,
       p.script_id                                            AS script_id
FROM parameters p
         INNER JOIN blocks b ON p.block_id = b.id
         INNER JOIN levels l ON b.level_id = l.id
         INNER JOIN models m ON l.model_id = m.id
         LEFT JOIN parameters pref on pref.id = p.ref_id
         LEFT JOIN boards ON m.board_id = boards.id
         INNER JOIN lwp ON b.level_id = lwp.id AND p.id = lwp.datapoint_id
WHERE b.type != 'VIRTUAL'
UNION
SELECT s.id                                                            AS id,
       s.name                                                          AS name,
       s.type                                                          AS type,
       s.size                                                          AS size,
       s.data_type                                                     AS data_type,
       s.ref_id                                                        AS ref_id,
       m.id                                                            AS model_id,
       case
           when m.board_id is NULL then m.name
           else boards.name
           end                                                         AS model_name,
       m.type                                                          AS model_type,
       s.block_id,
       case
           when lwp.path like concat(m.name, '/%') then RIGHT(lwp.path, LENGTH(lwp.path) - LENGTH(m.name) -1)
           else lwp.path end                                           AS block_path,
       b.name                                                          AS block_name,
       CONCAT_WS('/', NULLIF(lwp.path, ''), b.name, s.name)            AS path,
       ''                                                              AS alias,
       s.default_value                                                 AS default_value,
       NULL::text                                                      AS default_value_text,
       NULL::jsonb                                                     AS default_value_object,
       s.is_favorite                                                   AS is_favorite,
       s.is_visible                                                    AS is_visible,
       NULL::boolean                                                   AS is_tunable,
       case
           when s.ref_id is NULL then
               s.library_data::jsonb ->> 'description'
           else
               sref.library_data::jsonb ->> 'description'
           end
                                                                       AS description,
       null::boolean                                          AS is_read_only,
       case
           when s.ref_id is NULL then
               s.library_data::jsonb ->> 'unit'
           else
               sref.library_data::jsonb ->> 'unit'
           end
                                                              AS unit,
       case
           when s.ref_id is NULL then
               s.library_data::jsonb -> 'elementNames'
           else
               sref.library_data::jsonb -> 'elementNames'
           end
                                                              AS element_names,
       case
           when s.ref_id is NULL then
               s.library_data::jsonb -> 'enumValues'
           else
               sref.library_data::jsonb -> 'enumValues'
           end
                                                              AS enum_values,
       case
           when s.ref_id is NULL then
               (s.library_data::jsonb ->> 'folderId')::uuid
           else
               (sref.library_data::jsonb ->> 'folderId')::uuid
           end
                                                              AS folder_id,
       case
           when s.ref_id is NULL then
               (s.library_data::jsonb ->> 'position')::int
           else
               (sref.library_data::jsonb ->> 'position')::int
           end
                                                              AS pos,
       m.project_id                                           AS project_id,
       NULL::uuid                                             AS file_id,
       NULL::uuid                                             AS script_id
FROM signals s
         INNER JOIN blocks b ON s.block_id = b.id
         INNER JOIN levels l ON b.level_id = l.id
         INNER JOIN models m ON l.model_id = m.id
         LEFT JOIN signals sref on sref.id = s.ref_id
         LEFT JOIN boards ON m.board_id = boards.id
         INNER JOIN lwp ON b.level_id = lwp.id AND s.id = lwp.datapoint_id
WHERE b.type != 'VIRTUAL'
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints(datapoint_ids uuid[]) OWNER TO postgres;

--
-- TOC entry 949 (class 1255 OID 40122)
-- Name: datapoints_tree_by_model(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints_tree_by_model(lookup_model_id uuid) RETURNS TABLE(id uuid, name text, path text, parent_id uuid, type text, size integer[], default_value double precision[], default_value_text text, default_value_object jsonb, data_type text, is_favorite boolean, is_visible boolean, is_tunable boolean, file_id uuid, script_id uuid, ref_id uuid, model_id uuid, model_type text)
    LANGUAGE sql
    AS $$
WITH lwp AS (SELECT * FROM levels_with_path('/', (SELECT project_id FROM models WHERE id = lookup_model_id)))
SELECT lvl.id                   as id,
       lvl.name                 as name,
       lwp.path                 as path,
       lvl.parent_id            as parent_id,
       'LEVEL'                  as type,
       NULL::integer[]          as size,
       '{}'::double precision[] as default_value,
       NULL::text               as default_value_text,
       NULL::jsonb              as default_value_object,
       NULL                     as data_type,
       NULL::boolean            as is_favorite,
       NULL::boolean            as is_visible,
       NULL::boolean            as is_tunable,
       NULL::uuid               as file_id,
       NULL::uuid               as script_id,
       null::uuid               as ref_id,
       lookup_model_id			as model_id,
       NULL::text               as model_type
FROM levels lvl
    LEFT JOIN lwp
ON lvl.id = lwp.id
WHERE lwp.model_id = lookup_model_id
UNION
SELECT b.id                          as id,
       b.name                        as name,
       concat(lwp.path, '/', b.name) as path,
       b.level_id                    as parent_id,
       'BLOCK'                       as type,
       NULL::integer[]               as size,
       '{}'::double precision[]      as default_value,
       NULL::text                    as default_value_text,
       NULL::jsonb                   as default_value_object,
       NULL                          as data_type,
       NULL::boolean                 as is_favorite,
       NULL::boolean                 as is_visible,
       NULL::boolean                 as is_tunable,
       NULL::uuid                    as file_id,
       NULL::uuid                    as script_id,
       b.ref_id                      as ref_id,
       lookup_model_id			     as model_id,
       NULL::text                    as model_type
FROM blocks b
    LEFT JOIN lwp
ON b.level_id = lwp.id
WHERE lwp.model_id = lookup_model_id AND b.type != 'VIRTUAL'
UNION
SELECT p.id                                       AS id,
       p.name                                     AS name,
       concat(lwp.path, '/', b.name, '/', p.name) as path,
       case 
            when b.type = 'IO' or b.type = 'SUBSYSTEM' 
            then b.subsystem_level_id
            else b.id 
            end                                   as parent_id,
       'PARAMETER'                                AS type,
       p.size as size,
       p.default_value                            as default_value,
       p.default_value_text                       as default_value_text,
       p.default_value_object                     as default_value_object,
       p.data_type                                AS data_type,
       p.is_favorite                              as is_favorite,
       p.is_visible                               as is_visible,
       p.is_tunable                               as is_tunable,
       p.file_id                                  as file_id,
       p.script_id                                as script_id,
       p.ref_id                                   as ref_id,
       lookup_model_id				              as model_id,
       NULL::text                                 as model_type
FROM parameters p
    LEFT JOIN blocks b
ON p.block_id = b.id
    LEFT JOIN lwp ON b.level_id = lwp.id
WHERE lwp.model_id = lookup_model_id AND b.type != 'VIRTUAL'
UNION
SELECT s.id                                       AS id,
       s.name                                     AS name,
       concat(lwp.path, '/', b.name, '/', s.name) as path,
       case 
            when b.type = 'IO' or b.type = 'SUBSYSTEM' 
            then b.subsystem_level_id
            else b.id 
            end                                   as parent_id,
       s.type                                     as type,
       s.size as size,
       s.default_value                            as default_value,
       NULL::text                                 as default_value_text,
       NULL::jsonb                                as default_value_object,
       s.data_type                                AS data_type,
       s.is_favorite                              as is_favorite,
       s.is_visible                               as is_visible,
       NULL::boolean                              as is_tunable,
       NULL::uuid                                 as file_id,
       NULL::uuid                                 as script_id,
       s.ref_id                                   as ref_id,
       lookup_model_id			                  as model_id,
       NULL::text                                 as model_type
FROM signals s
    LEFT JOIN blocks b
ON s.block_id = b.id
    LEFT JOIN lwp ON b.level_id = lwp.id
WHERE lwp.model_id = lookup_model_id AND b.type != 'VIRTUAL'
ORDER BY path ASC
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints_tree_by_model(lookup_model_id uuid) OWNER TO postgres;

--
-- TOC entry 933 (class 1255 OID 39928)
-- Name: datapoints_tree_by_project(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints_tree_by_project(lookup_project_id uuid) RETURNS TABLE(id uuid, name text, path text, parent_id uuid, type text, size integer[], default_value double precision[], default_value_text text, default_value_object jsonb, data_type text, is_favorite boolean, is_visible boolean, is_tunable boolean, file_id uuid, script_id uuid, ref_id uuid, model_id uuid, model_type text)
    LANGUAGE sql
    AS $$
WITH lwp AS (SELECT * FROM levels_with_path('/', lookup_project_id))
SELECT lvl.id                   as id,
       lvl.name                 as name,
       lwp.path                 as path,
       lvl.parent_id            as parent_id,
       'LEVEL'                  as type,
       NULL::integer[]          as size,
       '{}'::double precision[] as default_value,
       NULL::text               as default_value_text,
       NULL::jsonb              as default_value_object,
       NULL::text               as data_type,
       NULL::boolean            as is_favorite,
       NULL::boolean            as is_visible,
       NULL::boolean            as is_tunable,
       NULL::uuid               as file_id,
       NULL::uuid               as script_id,
       null::uuid               as ref_id,
       lvl.model_id             as model_id,
       m."type"                 as model_type
FROM levels lvl
    LEFT JOIN lwp
ON lvl.id = lwp.id
    LEFT JOIN models m
ON lvl.model_id = m.id 
WHERE m.project_id = lookup_project_id
UNION
SELECT b.id                          as id,
       b.name                        as name,
       concat(lwp.path, '/', b.name) as path,
       b.level_id                    as parent_id,
       'BLOCK'                       as type,
       NULL::integer[]               as size,
       '{}'::double precision[]      as default_value,
       NULL::text                    as default_value_text,
       NULL::jsonb                   as default_value_object,
       NULL                          as data_type,
       NULL::boolean                 as is_favorite,
       NULL::boolean                 as is_visible,
       NULL::boolean                 as is_tunable,
       NULL::uuid                    as file_id,
       NULL::uuid                    as script_id,
       b.ref_id                      as ref_id,
       m.id                          as model_id,
       m."type"                      as model_type
FROM blocks b
    LEFT JOIN lwp
ON b.level_id = lwp.id
    LEFT JOIN models m
ON lwp.model_id = m.id 
WHERE m.project_id = lookup_project_id AND b.type != 'VIRTUAL'
UNION
SELECT p.id                                       AS id,
       p.name                                     AS name,
       concat(lwp.path, '/', b.name, '/', p.name) as path,
       case 
            when b.type = 'IO' or b.type = 'SUBSYSTEM' 
            then b.subsystem_level_id
            else b.id 
            end                                   as parent_id,
       'PARAMETER'                                AS type,
       p.size as size,
       p.default_value                            as default_value,
       p.default_value_text                       as default_value_text,
       p.default_value_object                     as default_value_object,
       p.data_type                                AS data_type,
       p.is_favorite                              as is_favorite,
       p.is_visible                               as is_visible,
       p.is_tunable                               as is_tunable,
       p.file_id                                  as file_id,
       p.script_id                                as script_id,
       p.ref_id                                   as ref_id,
       m.id                                       as model_id,
       m."type"                                   as model_type
FROM parameters p
    LEFT JOIN blocks b
ON p.block_id = b.id
    LEFT JOIN lwp 
ON b.level_id = lwp.id
    LEFT JOIN models m
ON lwp.model_id = m.id 
WHERE m.project_id = lookup_project_id AND b.type != 'VIRTUAL'
UNION
SELECT s.id                                       AS id,
       s.name                                     AS name,
       concat(lwp.path, '/', b.name, '/', s.name) as path,
       case 
            when b.type = 'IO' or b.type = 'SUBSYSTEM' 
            then b.subsystem_level_id
            else b.id 
            end                                   as parent_id,
       s.type                                     as type,
       s.size as size,
       s.default_value                            as default_value,
       NULL::text                                 as default_value_text,
       NULL::jsonb                                as default_value_object,
       s.data_type                                AS data_type,
       s.is_favorite                              as is_favorite,
       s.is_visible                               as is_visible,
       NULL::boolean                              as is_tunable,
       NULL::uuid                                 as file_id,
       NULL::uuid                                 as script_id,
       s.ref_id                                   as ref_id,
       m.id                                       as model_id,
       m."type"                                   as model_type
FROM signals s
    LEFT JOIN blocks b
ON s.block_id = b.id
    LEFT JOIN lwp 
ON b.level_id = lwp.id
    LEFT JOIN models m
ON lwp.model_id = m.id 
WHERE m.project_id = lookup_project_id AND b.type != 'VIRTUAL'
ORDER BY path ASC
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.datapoints_tree_by_project(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 925 (class 1255 OID 39734)
-- Name: delete_unreferenced_nodes(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.delete_unreferenced_nodes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM nodes
    WHERE id IN
          (SELECT node_id FROM modified_ports
           WHERE modified_ports.node_id NOT IN
                 (SELECT node_id FROM ports p WHERE p.node_id = ANY(SELECT node_id from modified_ports)));

    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.delete_unreferenced_nodes() OWNER TO postgres;

--
-- TOC entry 935 (class 1255 OID 39957)
-- Name: dynamic_datapoints(uuid, uuid, uuid, uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.dynamic_datapoints(lookup_block_id uuid, lookup_board_id uuid, lookup_model_id uuid, lookup_project_id uuid) RETURNS TABLE(id uuid, name text, type text, size integer[], data_type text, ref_id uuid, model_id uuid, model_name text, model_type text, block_id uuid, block_path text, block_name text, path text, alias text, default_value double precision[], default_value_text text, default_value_object jsonb, is_favorite boolean, is_visible boolean, is_tunable boolean, description text, is_read_only boolean, unit text, element_names jsonb, enum_values jsonb, folder_id uuid, pos integer, project_id uuid, file_id uuid, script_id uuid, metadata jsonb)
    LANGUAGE plpgsql
    AS $$
DECLARE
with_query text := 'WITH lwp AS (SELECT * FROM get_paths(%s, %s)) ';
    parameters_query
text := '
        SELECT
        p.id AS id,
        p.name AS name,
        ''PARAMETER'' AS type,
        p.size AS size,
        p.data_type AS data_type,
		p.ref_id,
        m.id AS model_id,
        m.name AS model_name,
        m.type AS model_type,
        p.block_id,
        CASE WHEN m.board_id IS NULL AND lwp.path LIKE ''%/%'' THEN RIGHT(lwp.path, LENGTH(lwp.path)-LENGTH(m.name)-1)
			 WHEN m.board_id IS NOT NULL THEN RIGHT(LEFT(lwp.path, LENGTH(lwp.path)-LENGTH(m.name)-1), -1)
             ELSE '''' END AS block_path,
        b.name AS block_name,
       case when m.board_id is NULL then CONCAT_WS(''/'', NULLIF(lwp.path, ''''),  NULLIF(b.name, ''''), p.name)
             else CONCAT_WS(''/'', RIGHT(lwp.path, LENGTH(lwp.path)-1), b.name, p.name) end as path,
        '''' as alias,
        p.default_value AS default_value,
        p.default_value_text as default_value_text,
        p.default_value_object as default_value_object,
        p.is_favorite AS is_favorite,
        p.is_visible AS is_visible,
        p.is_tunable AS is_tunable,
        pref.library_data::jsonb->>''description'' as description,
        p.is_read_only,
        pref.library_data::jsonb->>''unit'' as unit,
        pref.library_data::jsonb->''elementNames'' as element_names,
        pref.library_data::jsonb->''enumValues'' as enum_values,
        (pref.library_data::jsonb->>''folderId'')::uuid as folder_id,
        (pref.library_data::jsonb->>''position'')::int as pos,
        m.project_id as project_id,
        p.file_id as file_id,
        p.script_id as script_id,
        p.metadata as metadata
    FROM parameters p
        INNER JOIN blocks b ON p.block_id = b.id AND b.type != ''VIRTUAL''
        LEFT JOIN parameters pref on pref.id = p.ref_id ';
    signals_query
text := 'SELECT
        s.id AS id,
        s.name AS name,
        s.type AS type,
        s.size AS size,
        s.data_type AS data_type,
		s.ref_id,
        m.id AS model_id,
        m.name AS model_name,
        m.type AS model_type,
        s.block_id,
        CASE WHEN m.board_id IS NULL AND lwp.path LIKE ''%/%'' THEN RIGHT(lwp.path, LENGTH(lwp.path)-LENGTH(m.name)-1)
			 WHEN m.board_id IS NOT NULL THEN RIGHT(LEFT(lwp.path, LENGTH(lwp.path)-LENGTH(m.name)-1), -1)
             ELSE '''' END AS block_path,
        b.name AS block_name,
       case when m.board_id is NULL then CONCAT_WS(''/'', NULLIF(lwp.path, ''''),  NULLIF(b.name, ''''), s.name)
             else CONCAT_WS(''/'', RIGHT(lwp.path, LENGTH(lwp.path)-1), b.name, s.name) end as path,
        '''' as alias,
        s.default_value AS default_value,
        NULL::text as default_value_text,
        NULL::jsonb as default_value_object,
        s.is_favorite AS is_favorite,
        s.is_visible AS is_visible,
        NULL::boolean as is_tunable,
        sref.library_data::jsonb->>''description'' as description,
        NULL::boolean as is_read_only,
        sref.library_data::jsonb->>''unit'' as unit,
        sref.library_data::jsonb->''elementNames'' as element_names,
        sref.library_data::jsonb->''enumValues'' as enum_values,
        (sref.library_data::jsonb->>''folderId'')::uuid as folder_id,
        (sref.library_data::jsonb->>''position'')::int as pos,
        m.project_id as project_id,
        NULL::uuid as file_id,
        NULL::uuid as script_id,
        s.metadata as metadata
    FROM signals s
        INNER JOIN blocks b ON s.block_id = b.id AND b.type != ''VIRTUAL''
        LEFT JOIN signals sref on sref.id = s.ref_id ';
    common_joins
text := '
        INNER JOIN lwp ON b.level_id = lwp.level_id
        INNER JOIN models m ON lwp.model_id = m.id ';
    not_by_board_join
text := 'LEFT JOIN boards ON m.board_id = boards.id ';
BEGIN
    -- only one of the argument should be provided (XOR)
    IF
NOT ((((lookup_block_id IS NULL) != (lookup_board_id IS NULL)) != (lookup_model_id IS NULL)) !=
             (lookup_project_id IS NULL)) THEN
        RAISE EXCEPTION 'exactly one of {lookup_block_id, lookup_board_id, lookup_model_id, lookup_project_id} must be provided';
END IF;

    -- add joins common to both the parameter query and the signals query
    parameters_query
= concat(parameters_query, common_joins);
    signals_query
= concat(signals_query, common_joins);

    IF
lookup_block_id IS NOT NULL THEN -- add joins for datapoints by block

        with_query = format(with_query, 'null', format('(SELECT l.model_id FROM blocks b
            INNER JOIN levels l ON b.level_id = l.id WHERE b.id = %L)', lookup_block_id));

        parameters_query
= concat(parameters_query, not_by_board_join, format('WHERE b.id = %L ', lookup_block_id));

        signals_query
= concat(signals_query, not_by_board_join, format('WHERE b.id = %L;', lookup_block_id));

    ELSEIF
lookup_board_id IS NOT NULL THEN -- add joins for datapoints by board

        with_query = format(with_query, format('(SELECT b.project_id FROM boards b WHERE b.id = %L)', lookup_board_id),
                            'null');

        parameters_query
= concat(parameters_query, format('LEFT JOIN table_datapoints t ON t.parameter_id = p.id
            INNER JOIN boards ON m.board_id = boards.id OR t.board_id = boards.id WHERE boards.id = %L ',
                                              lookup_board_id));

        signals_query
= concat(signals_query, format('LEFT JOIN table_datapoints t ON t.signal_id = s.id
            INNER JOIN boards ON m.board_id = boards.id OR t.board_id = boards.id WHERE boards.id = %L;',
                                           lookup_board_id));

    ELSEIF
lookup_model_id IS NOT NULL THEN -- add joins for datapoints by model

        with_query = format(with_query, 'null', format('%L', lookup_model_id));

        parameters_query
= concat(parameters_query, not_by_board_join, format('WHERE m.id = %L ', lookup_model_id));

        signals_query
= concat(signals_query, not_by_board_join, format('WHERE m.id = %L;', lookup_model_id));

ELSE -- add joins for datapoints by project

        with_query = format(with_query, format('%L', lookup_project_id), 'null');

        parameters_query
=
                concat(parameters_query, not_by_board_join, format('WHERE m.project_id = %L ', lookup_project_id));

        signals_query
= concat(signals_query, not_by_board_join, format('WHERE m.project_id = %L;', lookup_project_id));
END IF;

    --raise warning 'Value: %', concat(with_query, parameters_query, 'UNION ', signals_query);

RETURN QUERY EXECUTE concat(with_query, parameters_query, 'UNION ', signals_query);
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.dynamic_datapoints(lookup_block_id uuid, lookup_board_id uuid, lookup_model_id uuid, lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 924 (class 1255 OID 39685)
-- Name: exec_block_cb(uuid, uuid, text, text); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.exec_block_cb(block_ref_id uuid, block_id uuid, opalpy_path text, param_name text) RETURNS text
    LANGUAGE plpython3u
    AS $_$
import json
import sys
import math
from inspect import getmembers, isfunction

sys.path.append(opalpy_path)
from opalpy import cb_util

global handle
handle = {}
handle["plpy"] = plpy
handle["block_id"] = block_id

globals_for_exec = {
    "handle": handle,
    "math": math,
}

cb_util_functions = {name: fn for name, fn in getmembers(cb_util, isfunction)}
globals_for_exec.update(cb_util_functions)


plan = plpy.prepare("SELECT callbacks FROM library_block_tree WHERE id = $1", ["uuid"])
callbacks = plpy.execute(plan, [block_ref_id])
if callbacks:
    get_callbacks = callbacks[0].get("callbacks")
    if get_callbacks:
        parsed_callbacks = json.loads(get_callbacks)
        for i in parsed_callbacks:
            if param_name in i.get("triggers"):
                callback = i.get("callback")
                exec(
                    callback,
                    globals_for_exec,
                )

$_$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.exec_block_cb(block_ref_id uuid, block_id uuid, opalpy_path text, param_name text) OWNER TO postgres;

--
-- TOC entry 923 (class 1255 OID 39683)
-- Name: get_active_parameter_set(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_active_parameter_set(parameter_id uuid) RETURNS uuid
    LANGUAGE sql
    AS $$
SELECT ps.id
FROM parameter_sets AS ps
LEFT JOIN models ON models.project_id = ps.project_id
LEFT JOIN levels ON levels.model_id  = models.id
LEFT JOIN blocks ON blocks.level_id = levels.id
LEFT JOIN parameters ON parameters.block_id = blocks.id
where
parameters.id = parameter_id AND
ps.type = (CASE
	WHEN models.state = 'RUNNING' OR models.state = 'PAUSED'
	THEN 'ONLINE'
	ELSE 'INITIAL'
END);
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_active_parameter_set(parameter_id uuid) OWNER TO postgres;

--
-- TOC entry 916 (class 1255 OID 39536)
-- Name: get_all_earlier_versions_schemas(integer); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_all_earlier_versions_schemas(current_version integer) RETURNS SETOF record
    LANGUAGE plpgsql
    AS $$
DECLARE
    rec record;
    version int;
BEGIN
    FOR rec IN SELECT nspname FROM pg_catalog.pg_namespace WHERE nspname LIKE 'workspace%'
        LOOP
            EXECUTE format('SELECT m.version from %I.migrations m order by m.id desc limit 1', rec.nspname) INTO version;
            IF current_version >= version THEN
                RETURN NEXT rec;
            END IF;
        END LOOP;
    RETURN;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_all_earlier_versions_schemas(current_version integer) OWNER TO postgres;

--
-- TOC entry 943 (class 1255 OID 40012)
-- Name: get_excluded_column_names_on_import(boolean); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_excluded_column_names_on_import(duplicate boolean) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
BEGIN
CASE duplicate
        WHEN true
            THEN return jsonb_build_object('migrations', jsonb_build_array(),
                                           'projects', json_build_array('path', 'checksum', 'created_at', 'modified_at'),
                                           'scripts', json_build_array('pid', 'execution_count', 'run_state', 'last_run_date', 'created_at', 'modified_at', 'logs', 'exception_type', 'exception_value'),
                                           'files', json_build_array('created_at', 'modified_at'),
                                           'boards', jsonb_build_array('created_at', 'modified_at'),
                                           'models', jsonb_build_array('state', 'created_at', 'modified_at'),
                                           'levels', jsonb_build_array('created_at', 'modified_at'),
                                           'blocks', jsonb_build_array('created_at', 'modified_at'),
                                           'parameters', jsonb_build_array('created_at', 'modified_at'),
                                           'signals', jsonb_build_array('created_at', 'modified_at'),
                                           'nodes', jsonb_build_array('created_at', 'modified_at'),
                                           'ports', jsonb_build_array('created_at', 'modified_at'),
                                           'links', jsonb_build_array('created_at', 'modified_at'),
                                           'aliases', jsonb_build_array('created_at', 'modified_at'),
                                           'parameter_sets', jsonb_build_array('created_at', 'modified_at'),
                                           'parameter_set_values', jsonb_build_array('created_at', 'modified_at'),
                                           'table_datapoints', jsonb_build_array('created_at'),
                                           'connections', jsonb_build_array('created_at', 'modified_at'),
                                           'simulation_configurations', jsonb_build_array('created_at', 'modified_at'),
                                           'io_assignments', jsonb_build_array('created_at', 'modified_at'));
ELSE return jsonb_build_object('migrations', jsonb_build_array(),
                                       'projects', json_build_array('path', 'checksum'),
                                       'scripts', json_build_array(),
                                       'files', json_build_array(),
                                       'boards', jsonb_build_array(),
                                       'models', jsonb_build_array('state'),
                                       'levels', jsonb_build_array(),
                                       'blocks', jsonb_build_array(),
                                       'parameters', jsonb_build_array(),
                                       'signals', jsonb_build_array(),
                                       'nodes', jsonb_build_array(),
                                       'ports', jsonb_build_array(),
                                       'links', jsonb_build_array(),
                                       'aliases', jsonb_build_array(),
                                       'parameter_sets', jsonb_build_array(),
                                       'parameter_set_values', jsonb_build_array(),
                                       'table_datapoints', jsonb_build_array(),
                                       'connections', jsonb_build_array(),
                                       'simulation_configurations', jsonb_build_array(),
                                       'io_assignments', jsonb_build_array());
END CASE;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_excluded_column_names_on_import(duplicate boolean) OWNER TO postgres;

--
-- TOC entry 942 (class 1255 OID 40011)
-- Name: get_export_queries(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_export_queries(lookup_project_id uuid) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
BEGIN
return jsonb_build_object('migrations', copyto_migrations_query(),
                          'projects', copyto_projects_query(lookup_project_id),
                          'scripts', copyto_scripts_query(lookup_project_id),
                          'files', copyto_files_query(lookup_project_id),
                          'boards', copyto_boards_query(lookup_project_id),
                          'models', copyto_models_query(lookup_project_id),
                          'levels', copyto_levels_query(lookup_project_id),
                          'blocks', copyto_blocks_query(lookup_project_id),
                          'parameters', copyto_parameters_query(lookup_project_id),
                          'signals', copyto_signals_query(lookup_project_id),
                          'nodes', copyto_nodes_query(lookup_project_id),
                          'aliases', copyto_aliases_query(lookup_project_id),
                          'ports', copyto_ports_query(lookup_project_id),
                          'links', copyto_links_query(lookup_project_id),
                          'parameter_sets', copyto_parameter_sets_query(lookup_project_id),
                          'parameter_set_values', copyto_parameter_set_values_query(lookup_project_id),
                          'table_datapoints', copyto_table_datapoints_query(lookup_project_id),
                          'connections', copyto_connections_query(lookup_project_id),
                          'simulation_configurations', copyto_simulation_configurations_query(lookup_project_id),
                          'io_assignments', copyto_io_assignments_query(lookup_project_id));
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_export_queries(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 944 (class 1255 OID 40050)
-- Name: get_list_param_values(uuid[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_list_param_values(ids uuid[]) RETURNS TABLE(id uuid, value double precision[], value_text text, value_object jsonb, value_expression text[])
    LANGUAGE sql
    AS $$
SELECT p.id AS id,
       CASE WHEN psv.value IS NULL
                THEN p.default_value
            ELSE psv.value END AS value,
       CASE WHEN psv.value_text IS NULL
                THEN p.default_value_text
            ELSE psv.value_text END AS value_text,
       CASE WHEN psv.value_object IS NULL
                THEN p.default_value_object
            ELSE psv.value_object END AS value_object,
        psv.value_expression as value_expression --There is no default value for value_expression
FROM parameters p
         LEFT JOIN blocks b ON b.id = p.block_id
         LEFT JOIN levels l ON l.id = b.level_id
         LEFT JOIN models m ON m.id = l.model_id
         LEFT JOIN parameter_sets ps ON m.project_id = ps.project_id
         LEFT JOIN parameter_set_values psv ON psv.parameter_set_id = ps.id and psv.parameter_id = p.id
WHERE p.id = ANY(ids)
  AND ps.type = (CASE WHEN m.state = 'RUNNING' or m.state = 'PAUSED' THEN 'ONLINE' ELSE 'INITIAL' END)
ORDER BY
    array_position(ids, p.id)
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_list_param_values(ids uuid[]) OWNER TO postgres;

--
-- TOC entry 903 (class 1255 OID 39076)
-- Name: get_lwp_all_libraries(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_lwp_all_libraries() RETURNS SETOF record
    LANGUAGE plpgsql
    AS $$
DECLARE
    rec record;
    p_id uuid;
BEGIN
    FOR p_id IN SELECT id FROM projects WHERE type='LIBRARY'
        LOOP
            FOR rec IN SELECT * FROM levels_with_path('/', p_id)
                LOOP
                    RETURN NEXT rec;
                END LOOP;
        END LOOP;
END $$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_lwp_all_libraries() OWNER TO postgres;

--
-- TOC entry 913 (class 1255 OID 39527)
-- Name: get_paths(uuid, uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_paths(lookup_project_id uuid, lookup_model_id uuid) RETURNS TABLE(level_id uuid, board_id uuid, path text, model_id uuid)
    LANGUAGE plpgsql
    AS $$
DECLARE
    levels_recursive text := 'WITH recursive cte_levels(level_id, board_id, path, model_id) AS (
        SELECT l.id, NULL::UUID, l.name, l.model_id
        FROM levels l JOIN models m ON m.id = l.model_id
        WHERE l.parent_id is null %s
    UNION ALL
        SELECT lvl.id, NULL::UUID,
        CONCAT_WS(''/'', NULLIF(cte_levels.path, ''''), NULLIF(lvl.name, '''')),
        lvl.model_id
        FROM levels lvl INNER JOIN cte_levels ON cte_levels.level_id = lvl.parent_id
    )';
    levels_select    text := 'SELECT * FROM cte_levels';
    boards_recursive text := 'cte_boards(level_id, board_id, path, model_id) AS (
        SELECT l.id, b.id, b.name, m.id
        FROM boards b
            LEFT JOIN models m ON m.board_id = b.id
            LEFT JOIN levels l ON l.model_id = m.id
        WHERE b.parent_id IS NULL %s
    UNION ALL
        SELECT l.id, b.id,
        CONCAT_WS(''/'', NULLIF(cte_boards.path, ''''), NULLIF(b.name, '''')),
        m.id
        FROM boards b INNER JOIN cte_boards ON cte_boards.board_id = b.parent_id
            LEFT JOIN models m ON m.board_id = b.id
            LEFT JOIN levels l ON l.model_id = m.id
    )';
    board_select     text := 'SELECT level_id, board_id, CONCAT(''/'', path) AS path, model_id FROM cte_boards';
    query            text;
BEGIN
    IF NOT ((lookup_project_id IS NOT NULL) != (lookup_model_id IS NOT NULL)) THEN -- XOR
        RAISE EXCEPTION 'exactly one of {lookup_project_id, lookup_model_id} must be provided';
    END IF;

    IF lookup_project_id IS NOT NULL THEN
        query = format('%s, %s %s UNION %s;',
                       format(levels_recursive,
                              format('AND m.project_id = %L AND m.board_id IS NULL', lookup_project_id)),
                       format(boards_recursive, format('AND b.project_id = %L ', lookup_project_id)),
                       levels_select, board_select);
    ELSE
        query = format('%s, %s %s UNION %s;',
                       format(levels_recursive, format('AND l.model_id = %L AND m.board_id IS NULL ', lookup_model_id)),
                       format(boards_recursive, ''),
                       levels_select, board_select, format(' WHERE m.id = %L ', lookup_model_id));
    END IF;

    --raise warning 'Value: %', query;

    RETURN QUERY EXECUTE query;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_paths(lookup_project_id uuid, lookup_model_id uuid) OWNER TO postgres;

--
-- TOC entry 941 (class 1255 OID 40010)
-- Name: get_table_names(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.get_table_names() RETURNS text[]
    LANGUAGE plpgsql
    AS $$
BEGIN
return '{
        "migrations",
		"projects",
		"scripts",
		"files",
		"boards",
		"models",
		"levels",
		"blocks",
		"parameters",
		"signals",
		"nodes",
		"ports",
		"links",
        "aliases",
		"parameter_sets",
		"parameter_set_values",
		"table_datapoints",
		"connections",
		"simulation_configurations",
        "io_assignments"
    }';
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.get_table_names() OWNER TO postgres;

--
-- TOC entry 917 (class 1255 OID 39643)
-- Name: levels_with_path(character, uuid[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path(separator character, level_ids uuid[]) RETURNS TABLE(id uuid, path text, model_id uuid)
    LANGUAGE sql
    AS $$
WITH recursive cte(origin, id, parent_id, path, model_id, model_name) AS (SELECT l.id as origin,
                                                                                 l.id,
                                                                                 l.parent_id,
                                                                                 l.name,
                                                                                 l.model_id,
                                                                                 m.name
                                                                          FROM levels l
                                                                                   INNER JOIN UNNEST(level_ids) AS x(id) ON l.id = x.id
                                                                                   INNER JOIN models m ON m.id = l.model_id
                                                                                   WHERE m."type" != 'PANEL'
                                                                          UNION ALL
                                                                          SELECT cte.origin                            as origin,
                                                                                 lvl.id,
                                                                                 lvl.parent_id,
                                                                                 concat(lvl.name, separator, cte.path) as path,
                                                                                 lvl.model_id,
                                                                                 cte.model_name
                                                                          FROM levels lvl
                                                                                   INNER JOIN cte ON cte.parent_id = lvl.id),
    cte_boards(origin, id, parent_id, path, model_id, model_name, board_id) AS (SELECT l.id as origin, 
                                                                                l.id, 
                                                                                l.parent_id, 
                                                                                b.name, 
                                                                                l.model_id, 
                                                                                m.name, 
                                                                                b.id
                                                                            FROM boards b
                                                                                LEFT JOIN models m ON m.board_id = b.id
                                                                                LEFT JOIN levels l ON l.model_id = m.id
                                                                                WHERE b.parent_id IS NULL
                                                                            UNION ALL
                                                                            SELECT l.id as origin, 
                                                                                    l.id, 
                                                                                    l.parent_id,
                                                                                    CONCAT_WS(separator, NULLIF(cte_boards.path, ''''), NULLIF(b.name, '''')),
                                                                                    m.id, 
                                                                                    m.name, 
                                                                                    b.id
                                                                            FROM boards b 
                                                                                INNER JOIN cte_boards ON cte_boards.board_id = b.parent_id
                                                                            LEFT JOIN models m ON m.board_id = b.id
                                                                            LEFT JOIN levels l ON l.model_id = m.id)
SELECT origin as id, path, model_id
FROM cte
WHERE path LIKE CONCAT(model_name, '%')
   OR path = ''
   OR path IS NULL
UNION 
SELECT origin as id, path, model_id FROM cte_boards
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path(separator character, level_ids uuid[]) OWNER TO postgres;

--
-- TOC entry 912 (class 1255 OID 39495)
-- Name: levels_with_path(character, uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path(separator character, lookup_project_id uuid) RETURNS TABLE(id uuid, path text, model_id uuid)
    LANGUAGE sql
    AS $$
WITH recursive cte(id, path, model_id) AS (
    SELECT levels.id, levels.name, levels.model_id
    FROM levels
             JOIN models m ON m.id = model_id
    WHERE parent_id is null AND m.project_id = lookup_project_id
    UNION ALL
    SELECT
        lvl.id,
        CONCAT_WS(separator, NULLIF(cte.path, ''), NULLIF(lvl.name, '')),
        lvl.model_id
    FROM levels lvl JOIN cte ON cte.id = lvl.parent_id
)
SELECT * FROM cte
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path(separator character, lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 926 (class 1255 OID 39743)
-- Name: levels_with_path_by_model(character, uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path_by_model(separator character, lookup_model_id uuid) RETURNS TABLE(id uuid, path text)
    LANGUAGE sql
    AS $$
WITH recursive cte(id, path) AS (
    SELECT levels.id, ''  -- root name is left as an empty string
    FROM levels
    WHERE parent_id is null AND model_id = lookup_model_id
    UNION ALL
    SELECT
        lvl.id,
        CONCAT_WS(separator, NULLIF(cte.path, ''), NULLIF(lvl.name, ''))
    FROM levels lvl JOIN cte ON cte.id = lvl.parent_id
)
SELECT * FROM cte
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.levels_with_path_by_model(separator character, lookup_model_id uuid) OWNER TO postgres;

--
-- TOC entry 939 (class 1255 OID 39966)
-- Name: list_assigned_files(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.list_assigned_files(lookup_project_id uuid) RETURNS uuid[]
    LANGUAGE plpgsql
    AS $$
DECLARE ids uuid[];
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    ids = ARRAY(SELECT DISTINCT f.id FROM files f
        JOIN parameters p on p.data_type = 'FILE' AND p.file_id = f.id
        JOIN block_ids_with_refs bir ON bir.id = p.block_id);
    return ids;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.list_assigned_files(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 938 (class 1255 OID 39965)
-- Name: list_assigned_workspace_files(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.list_assigned_workspace_files(lookup_project_id uuid) RETURNS uuid[]
    LANGUAGE plpgsql
    AS $$
DECLARE ids uuid[];
BEGIN
    PERFORM create_table_block_ids_with_refs(lookup_project_id);
    ids = ARRAY(SELECT DISTINCT f.id FROM files f
        JOIN parameters p on p.data_type = 'FILE' AND p.file_id = f.id
        JOIN block_ids_with_refs bir ON bir.id = p.block_id
                            WHERE f.project_id IS NULL);
    return ids;

END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.list_assigned_workspace_files(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 940 (class 1255 OID 39967)
-- Name: list_related_files(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.list_related_files(lookup_project_id uuid) RETURNS uuid[]
    LANGUAGE plpgsql
    AS $$
DECLARE ids uuid[];
BEGIN
    ids = ARRAY(SELECT f.id FROM files f WHERE f.project_id = lookup_project_id) || list_assigned_files(lookup_project_id);
    ids = ARRAY(SELECT DISTINCT * from UNNEST(ids));
    return ids;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.list_related_files(lookup_project_id uuid) OWNER TO postgres;

--
-- TOC entry 948 (class 1255 OID 40109)
-- Name: natural_sort_hierarchical(text); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.natural_sort_hierarchical(text) RETURNS text[]
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
    input_value text;
    path_parts text[];
    result text[];
    part text;
    normalized_part text;
    i integer;
    digit_match text;
    padded_number text;
    arr_length integer;
BEGIN
    input_value := $1;

    IF input_value IS NULL OR input_value = '[0]' THEN
        RETURN ARRAY['1'];
    ELSE
        result := ARRAY['0'];
    END IF;

    path_parts := string_to_array(input_value, '/');
    arr_length := array_length(path_parts, 1);

    IF arr_length IS NULL THEN
        result := array_append(result, input_value);
    RETURN result;
    END IF;

    FOR i IN 1..arr_length LOOP
            part := path_parts[i];

            normalized_part := part;

            BEGIN
                FOR digit_match IN SELECT (regexp_matches(part, '(\d+)', 'g'))[1] LOOP
                                padded_number := lpad(digit_match, 10, '0');
                normalized_part := regexp_replace(normalized_part, digit_match, padded_number, 'g');
                END LOOP;
            EXCEPTION WHEN OTHERS THEN
                        normalized_part := part;
            END;

        result := array_append(result, normalized_part);
        result := array_append(result, length(part)::text);
    END LOOP;

    result := array_append(result, '');
    result := array_append(result, arr_length::text);

RETURN result;
END;
$_$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.natural_sort_hierarchical(text) OWNER TO postgres;

--
-- TOC entry 922 (class 1255 OID 39680)
-- Name: notify_parameter_set_values(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_parameter_set_values() RETURNS trigger
    LANGUAGE plpython3u
    AS $$
# For more info on the use of TD instead of regular arguments, see https://www.postgresql.org/docs/current/plpython-trigger.html
edited_table = TD["args"][0]  # inserted or updated
# The payload in pg_notify has a max size of 8000 byte so we send events by chunk of 100 ids
CHUNK_SIZE = 100
rows = plpy.execute(f'SELECT parameter_id FROM {edited_table}')
parameters_chunks = [rows[i:i+CHUNK_SIZE] for i in range(0, len(rows), CHUNK_SIZE)]
for parameters_chunk in parameters_chunks:
    ids = ",".join([parameter["parameter_id"] for parameter in parameters_chunk])
    plpy.execute(f"""SELECT pg_notify('parameter_set_values:{edited_table}'::text, {plpy.quote_literal(ids)})""")
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.notify_parameter_set_values() OWNER TO postgres;

--
-- TOC entry 921 (class 1255 OID 39660)
-- Name: notify_projects_deleted(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_deleted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    channel text := 'projects:auto-export:deleted';
    project_id text;
BEGIN
    FOR project_id IN
        SELECT DISTINCT id FROM deleted
    LOOP
	    PERFORM pg_notify(channel, project_id);
    END LOOP;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_deleted() OWNER TO postgres;

--
-- TOC entry 920 (class 1255 OID 39658)
-- Name: notify_projects_inserted(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_inserted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    channel text := 'projects:auto-export:inserted';
    project_id text;
BEGIN
    FOR project_id IN
        SELECT DISTINCT id FROM inserted
    LOOP
	    PERFORM pg_notify(channel, project_id);
    END LOOP;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_inserted() OWNER TO postgres;

--
-- TOC entry 919 (class 1255 OID 39656)
-- Name: notify_projects_updated(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_updated() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    channel text := 'projects:auto-export:updated';
    project_id text;
BEGIN
    FOR project_id IN
        SELECT DISTINCT id FROM updated
    LOOP
	    PERFORM pg_notify(channel, project_id);
    END LOOP;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_updated() OWNER TO postgres;

--
-- TOC entry 945 (class 1255 OID 40051)
-- Name: parameters_tree_by_parameter_set(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.parameters_tree_by_parameter_set(parameterset_id uuid) RETURNS TABLE(id uuid, name text, path text, parent_id uuid, type text, value double precision[], value_text text, value_object jsonb, value_expression text[])
    LANGUAGE sql
    AS $$
WITH lwp AS (SELECT * FROM levels_with_path('/', (SELECT project_id FROM parameter_sets WHERE id = parameterset_id)))
SELECT
    p.id AS id,
    p.name AS name,
    concat(lwp.path, '/', b.name, '/', p.name) as path,
    b.id as parent_id,
    'PARAMETER' AS type,
    psv.value,
    psv.value_text,
    psv.value_object,
    psv.value_expression
FROM parameter_set_values psv, parameters p
                                   LEFT JOIN blocks b ON p.block_id = b.id
                                   LEFT JOIN lwp ON b.level_id = lwp.id
                                   LEFT JOIN models m ON lwp.model_id = m.id
                                   LEFT JOIN parameter_sets ps ON ps.project_id = m.project_id
WHERE psv.parameter_set_id = ps.id AND psv.parameter_id = p.id AND ps.id = parameterset_id AND b.type != 'VIRTUAL'
UNION
SELECT
    b.id AS id,
    b.name AS name,
    concat(lwp.path, '/', b.name) AS path,
    b.level_id AS parent_id,
    'BLOCK' AS type,
    '{}',
    null,
    null,
    null
FROM blocks b
         LEFT JOIN lwp ON b.level_id = lwp.id
         LEFT JOIN models m ON lwp.model_id = m.id
         LEFT JOIN parameter_sets ps ON ps.project_id = m.project_id
WHERE ps.id = parameterset_id AND b.type != 'VIRTUAL'
UNION
SELECT
    lvl.id AS id,
    lvl.name AS name,
    lwp.path AS path,
    lvl.parent_id AS parent_id,
    'LEVEL' AS type,
    '{}',
    null,
    null,
    null
FROM levels lvl
         LEFT JOIN lwp ON lvl.id = lwp.id
         LEFT JOIN models m ON lwp.model_id = m.id
         LEFT JOIN parameter_sets ps ON ps.project_id = m.project_id
WHERE ps.id = parameterset_id
ORDER BY path, name
    $$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.parameters_tree_by_parameter_set(parameterset_id uuid) OWNER TO postgres;

--
-- TOC entry 904 (class 1255 OID 39086)
-- Name: resolve_datapoints_ids(uuid[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.resolve_datapoints_ids(datapoints_ids uuid[]) RETURNS TABLE(project_id uuid, parameter_id uuid, signal_id uuid)
    LANGUAGE sql
    AS $$
SELECT models.project_id, parameter_id, signal_id
FROM (
         (SELECT block_id, p.id AS datapoint_id, p.id AS parameter_id, null AS signal_id FROM parameters p)
         UNION
         (SELECT block_id, s.id AS datapoint_id, null AS parameter_id, s.id AS signal_id FROM signals s)
     ) AS a
         LEFT JOIN blocks b ON a.block_id = b.id
         LEFT JOIN levels l ON b.level_id = l.id
         LEFT JOIN models ON l.model_id = models.id
WHERE datapoint_id = ANY(datapoints_ids);
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.resolve_datapoints_ids(datapoints_ids uuid[]) OWNER TO postgres;

--
-- TOC entry 946 (class 1255 OID 40052)
-- Name: set_index_value_expressions(text[], integer[], text[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.set_index_value_expressions(INOUT p_array_value_expressions text[], p_indexes integer[], p_value_expressions text[]) RETURNS text[]
    LANGUAGE plpgsql
    AS $$
DECLARE
i int = 1;
    idx int;
BEGIN
    FOREACH idx IN ARRAY p_indexes LOOP
            p_array_value_expressions[idx] := p_value_expressions[i];
            i := i + 1;
END LOOP;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.set_index_value_expressions(INOUT p_array_value_expressions text[], p_indexes integer[], p_value_expressions text[]) OWNER TO postgres;

--
-- TOC entry 905 (class 1255 OID 39087)
-- Name: set_index_values(double precision[], integer[], double precision[]); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.set_index_values(INOUT p_array double precision[], p_indexes integer[], p_values double precision[]) RETURNS double precision[]
    LANGUAGE plpgsql
    AS $$
DECLARE
    i int = 1;
    idx int;
BEGIN
    FOREACH idx IN ARRAY p_indexes LOOP
            p_array[idx] := p_values[i];
            i := i + 1;
        END LOOP;
END
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.set_index_values(INOUT p_array double precision[], p_indexes integer[], p_values double precision[]) OWNER TO postgres;

--
-- TOC entry 906 (class 1255 OID 39088)
-- Name: update_block_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE blocks AS b SET modified_at = current_timestamp FROM old_table AS o WHERE b.id = o.block_id;
    ELSE
        UPDATE blocks AS b SET modified_at = n.modified_at FROM new_table AS n WHERE b.id = n.block_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp() OWNER TO postgres;

--
-- TOC entry 907 (class 1255 OID 39089)
-- Name: update_board_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE boards AS b SET modified_at = current_timestamp FROM old_table AS o WHERE b.id = o.board_id;
    ELSE
        UPDATE boards AS b SET modified_at = current_timestamp FROM new_table AS n WHERE b.id = n.board_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp() OWNER TO postgres;

--
-- TOC entry 908 (class 1255 OID 39090)
-- Name: update_level_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE levels AS l SET modified_at = current_timestamp FROM old_table AS o WHERE l.id = o.level_id;
    ELSE
        UPDATE levels AS l SET modified_at = n.modified_at FROM new_table AS n WHERE l.id = n.level_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp() OWNER TO postgres;

--
-- TOC entry 909 (class 1255 OID 39091)
-- Name: update_model_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_model_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE models AS m SET modified_at = current_timestamp FROM old_table AS o WHERE m.id = o.model_id;
    ELSE
        UPDATE models AS m SET modified_at = n.modified_at FROM new_table AS n WHERE m.id = n.model_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_model_timestamp() OWNER TO postgres;

--
-- TOC entry 918 (class 1255 OID 39648)
-- Name: update_parameter_and_parameter_set_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_parameter_and_parameter_set_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE parameter_sets AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.parameter_set_id;
        UPDATE parameters AS pa SET modified_at = current_timestamp FROM old_table AS o WHERE pa.id = o.parameter_id;
    ELSE
        UPDATE parameter_sets AS p SET modified_at = n.modified_at FROM new_table AS n WHERE p.id = n.parameter_set_id;
        UPDATE parameters AS pa SET modified_at = n.modified_at FROM new_table AS n WHERE pa.id = n.parameter_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_parameter_and_parameter_set_timestamp() OWNER TO postgres;

--
-- TOC entry 910 (class 1255 OID 39093)
-- Name: update_project_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE projects AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.project_id;
    ELSE
        UPDATE projects AS p SET modified_at = n.modified_at FROM new_table AS n WHERE p.id = n.project_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp() OWNER TO postgres;

--
-- TOC entry 927 (class 1255 OID 39748)
-- Name: update_simulation_configuration_timestamp(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_simulation_configuration_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        UPDATE simulation_configurations AS sc SET modified_at = current_timestamp FROM old_table AS o WHERE sc.id = o.simulation_configuration_id;
    ELSE
        UPDATE simulation_configurations AS sc SET modified_at = n.modified_at FROM new_table AS n WHERE sc.id = n.simulation_configuration_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_simulation_configuration_timestamp() OWNER TO postgres;

--
-- TOC entry 915 (class 1255 OID 39533)
-- Name: update_slot_usage(); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_slot_usage() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        PERFORM update_slot_usage_for_node(OLD.simulation_node_id);
    ELSE
        PERFORM update_slot_usage_for_node(NEW.simulation_node_id);
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_slot_usage() OWNER TO postgres;

--
-- TOC entry 914 (class 1255 OID 39532)
-- Name: update_slot_usage_for_node(uuid); Type: FUNCTION; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE FUNCTION workspace_s1v8h2yrq9x15u1x.update_slot_usage_for_node(node_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    in_use integer;
    total integer;
BEGIN
    in_use = (SELECT count(*) FROM processing_slots WHERE processing_slots."type" = 'RT' AND processing_slots.simulation_node_id = node_id AND processing_slots.status = 'INUSE');
    total =  (SELECT count(*) FROM processing_slots WHERE processing_slots."type" = 'RT' AND processing_slots.simulation_node_id = node_id);

    UPDATE simulation_nodes AS sn
        SET slot_usage = jsonb_build_object('inUse', in_use, 'total', total)
        WHERE sn.id = node_id;
    RETURN true;
END;
$$;


ALTER FUNCTION workspace_s1v8h2yrq9x15u1x.update_slot_usage_for_node(node_id uuid) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 409 (class 1259 OID 39973)
-- Name: aliases; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.aliases (
    id uuid NOT NULL,
    project_id uuid NOT NULL,
    parameter_id uuid,
    signal_id uuid,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    CONSTRAINT chk_parameter_xor_signal CHECK ((((parameter_id IS NOT NULL) AND (signal_id IS NULL)) OR ((parameter_id IS NULL) AND (signal_id IS NOT NULL))))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.aliases OWNER TO postgres;

--
-- TOC entry 383 (class 1259 OID 38832)
-- Name: blocks; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.blocks (
    id uuid NOT NULL,
    level_id uuid,
    ref_id uuid,
    type text DEFAULT ''::text NOT NULL,
    name text DEFAULT ''::text,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    location point,
    angle double precision DEFAULT 0.0,
    is_label_visible boolean DEFAULT false,
    comment_state text DEFAULT 'NONE'::text,
    flip_state text DEFAULT 'NONE'::text,
    dimensions text DEFAULT ''::text,
    metadata jsonb DEFAULT '{}'::jsonb,
    library_data jsonb DEFAULT '{}'::jsonb,
    subsystem_level_id uuid,
    z_order integer,
    is_snapped_to_grid boolean DEFAULT true,
    is_ehs boolean DEFAULT false,
    is_phasor boolean DEFAULT false,
    is_emt boolean DEFAULT false,
    is_locked boolean DEFAULT false,
    time_step double precision DEFAULT 0.0,
    group_id uuid,
    label_location text,
    CONSTRAINT subsystem_level_id_null CHECK ((((subsystem_level_id IS NULL) AND ((type <> 'SUBSYSTEM'::text) OR (type <> 'IO'::text))) OR (((type = 'SUBSYSTEM'::text) OR (type = 'IO'::text)) AND (subsystem_level_id IS NOT NULL)))),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.blocks OWNER TO postgres;

--
-- TOC entry 384 (class 1259 OID 38850)
-- Name: boards; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.boards (
    id uuid NOT NULL,
    project_id uuid,
    parent_id uuid,
    type text DEFAULT 'TABLE'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    description text DEFAULT ''::text,
    z_order integer DEFAULT 0 NOT NULL,
    view_options jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.boards OWNER TO postgres;

--
-- TOC entry 385 (class 1259 OID 38862)
-- Name: connections; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.connections (
    id uuid NOT NULL,
    project_id uuid NOT NULL,
    parameter_id uuid,
    from_signal_id uuid,
    to_signal_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    is_enabled boolean DEFAULT true,
    index_from integer[],
    index_to integer[],
    type text DEFAULT 'USER'::text
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.connections OWNER TO postgres;

--
-- TOC entry 411 (class 1259 OID 40053)
-- Name: errors; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.errors (
    id uuid NOT NULL,
    project_id uuid,
    parameter_id uuid,
    parameter_set_id uuid,
    parameter_raw_index integer,
    severity_level text NOT NULL,
    code integer NOT NULL,
    message text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.errors OWNER TO postgres;

--
-- TOC entry 404 (class 1259 OID 39537)
-- Name: files; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.files (
    id uuid NOT NULL,
    project_id uuid,
    content bytea,
    name text NOT NULL,
    mime_type text NOT NULL,
    size integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.files OWNER TO postgres;

--
-- TOC entry 386 (class 1259 OID 38886)
-- Name: hubs; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.hubs (
    id uuid NOT NULL,
    type text DEFAULT 'REMOTE'::text NOT NULL,
    status text DEFAULT 'OFFLINE'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    description text DEFAULT ''::text,
    ip inet,
    port integer,
    os text,
    dolphin_networks jsonb,
    CONSTRAINT status_not_empty CHECK ((status <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.hubs OWNER TO postgres;

--
-- TOC entry 403 (class 1259 OID 39513)
-- Name: info; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.info (
    id uuid NOT NULL,
    name text DEFAULT ''::text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.info OWNER TO postgres;

--
-- TOC entry 406 (class 1259 OID 39775)
-- Name: io_assignments; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.io_assignments (
    id uuid NOT NULL,
    io_level_id uuid,
    model_level_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.io_assignments OWNER TO postgres;

--
-- TOC entry 387 (class 1259 OID 38898)
-- Name: levels; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.levels (
    id uuid NOT NULL,
    model_id uuid,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    scale double precision DEFAULT 1.2,
    metadata jsonb DEFAULT '{}'::jsonb,
    "position" point,
    background_id uuid
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.levels OWNER TO postgres;

--
-- TOC entry 388 (class 1259 OID 38910)
-- Name: links; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.links (
    id uuid NOT NULL,
    level_id uuid,
    from_port_id uuid,
    to_port_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    points path,
    is_open_ended boolean DEFAULT false,
    is_label_visible boolean DEFAULT false
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.links OWNER TO postgres;

--
-- TOC entry 382 (class 1259 OID 38828)
-- Name: migrations; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.migrations (
    id integer NOT NULL,
    version bigint,
    created_at timestamp with time zone
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.migrations OWNER TO postgres;

--
-- TOC entry 381 (class 1259 OID 38827)
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE SEQUENCE workspace_s1v8h2yrq9x15u1x.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE workspace_s1v8h2yrq9x15u1x.migrations_id_seq OWNER TO postgres;

--
-- TOC entry 5578 (class 0 OID 0)
-- Dependencies: 381
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER SEQUENCE workspace_s1v8h2yrq9x15u1x.migrations_id_seq OWNED BY workspace_s1v8h2yrq9x15u1x.migrations.id;


--
-- TOC entry 389 (class 1259 OID 38922)
-- Name: models; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.models (
    id uuid NOT NULL,
    project_id uuid,
    board_id uuid,
    type text DEFAULT 'STANDARD'::text NOT NULL,
    state text DEFAULT 'UNREACHABLE'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    description text DEFAULT ''::text,
    time_step double precision DEFAULT 0.0,
    is_grid_visible boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    solver_settings jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT board_id_when_panel_type CHECK ((((board_id IS NOT NULL) AND (type = 'PANEL'::text)) OR ((board_id IS NULL) AND (type <> 'PANEL'::text)))),
    CONSTRAINT state_not_empty CHECK ((state <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.models OWNER TO postgres;

--
-- TOC entry 405 (class 1259 OID 39703)
-- Name: nodes; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.nodes (
    id uuid NOT NULL,
    model_id uuid,
    name text,
    type text DEFAULT 'POWER'::text,
    data_type text DEFAULT 'REAL'::text,
    size integer[],
    value integer,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.nodes OWNER TO postgres;

--
-- TOC entry 390 (class 1259 OID 38937)
-- Name: parameter_set_values; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.parameter_set_values (
    id uuid NOT NULL,
    parameter_set_id uuid,
    parameter_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    value double precision[],
    value_text text,
    value_object jsonb,
    value_expression text[]
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.parameter_set_values OWNER TO postgres;

--
-- TOC entry 391 (class 1259 OID 38944)
-- Name: parameter_sets; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.parameter_sets (
    id uuid NOT NULL,
    project_id uuid,
    type text DEFAULT 'CUSTOM'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    description text DEFAULT ''::text,
    control_id integer,
    z_order integer DEFAULT 0,
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.parameter_sets OWNER TO postgres;

--
-- TOC entry 392 (class 1259 OID 38955)
-- Name: parameters; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.parameters (
    id uuid NOT NULL,
    block_id uuid,
    ref_id uuid,
    data_type text DEFAULT 'REAL'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    default_value double precision[],
    size integer[],
    is_displayed_on_canvas boolean DEFAULT false,
    is_favorite boolean DEFAULT false,
    is_visible boolean DEFAULT false,
    is_equal_mode boolean DEFAULT false,
    is_tunable boolean DEFAULT true,
    is_read_only boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    library_data jsonb DEFAULT '{}'::jsonb,
    file_id uuid,
    default_value_text text,
    is_ehs boolean DEFAULT false,
    is_phasor boolean DEFAULT false,
    is_emt boolean DEFAULT false,
    default_value_object jsonb,
    script_id uuid,
    CONSTRAINT datatype_not_empty CHECK ((data_type <> ''::text)),
    CONSTRAINT file_id_or_value_null CHECK ((((file_id IS NULL) AND (data_type <> 'FILE'::text)) OR (default_value = '{}'::double precision[]))),
    CONSTRAINT parameter_value_data_type_check CHECK ((((default_value_text IS NOT NULL) AND (data_type = 'STRING'::text)) OR ((default_value_object IS NOT NULL) AND (data_type = 'OBJECT'::text)) OR (default_value IS NOT NULL))),
    CONSTRAINT script_id_null_if_not_script_type CHECK ((((script_id IS NULL) AND (data_type <> 'SCRIPT'::text)) OR (data_type = 'SCRIPT'::text)))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.parameters OWNER TO postgres;

--
-- TOC entry 393 (class 1259 OID 38973)
-- Name: ports; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.ports (
    id uuid NOT NULL,
    block_id uuid,
    ref_id uuid,
    default_type text DEFAULT 'POWER'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    direction text DEFAULT 'NONE'::text NOT NULL,
    default_data_type text DEFAULT 'NONE'::text NOT NULL,
    alignment text DEFAULT ''::text NOT NULL,
    link_direction text DEFAULT 'ANY'::text NOT NULL,
    default_size integer[],
    is_ehs boolean DEFAULT false,
    is_phasor boolean DEFAULT false,
    is_emt boolean DEFAULT false,
    associated_port_id uuid,
    node_id uuid,
    metadata jsonb,
    library_data jsonb,
    is_enabled boolean DEFAULT true NOT NULL,
    CONSTRAINT link_direction_not_empty CHECK ((link_direction <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((default_type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.ports OWNER TO postgres;

--
-- TOC entry 394 (class 1259 OID 38986)
-- Name: processes; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.processes (
    id uuid NOT NULL,
    simulation_id uuid NOT NULL,
    processing_slot_id uuid,
    type text NOT NULL,
    started_at timestamp with time zone DEFAULT now(),
    ended_at timestamp with time zone,
    name text,
    path text,
    pid integer,
    simulation_critical boolean,
    exit_code integer,
    parameters text DEFAULT '[]'::text,
    is_managed boolean DEFAULT false,
    status text DEFAULT 'TERMINATED'::text,
    modified_at timestamp with time zone DEFAULT now(),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.processes OWNER TO postgres;

--
-- TOC entry 395 (class 1259 OID 38993)
-- Name: processing_slots; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.processing_slots (
    id uuid NOT NULL,
    simulation_node_id uuid NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'FREE'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    CONSTRAINT status_not_empty CHECK ((status <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.processing_slots OWNER TO postgres;

--
-- TOC entry 396 (class 1259 OID 39001)
-- Name: projects; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.projects (
    id uuid NOT NULL,
    type text DEFAULT 'STANDARD'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    version text DEFAULT '1'::text,
    path text DEFAULT ''::text,
    metadata jsonb DEFAULT '{}'::jsonb,
    exported_at timestamp with time zone,
    checksum text,
    is_auto_export_enabled boolean DEFAULT false,
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.projects OWNER TO postgres;

--
-- TOC entry 397 (class 1259 OID 39013)
-- Name: reservations; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.reservations (
    simulation_id uuid NOT NULL,
    processing_slot_id uuid NOT NULL,
    dolphin_link jsonb,
    interface_block_id uuid NOT NULL
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.reservations OWNER TO postgres;

--
-- TOC entry 407 (class 1259 OID 39808)
-- Name: scripts; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.scripts (
    id uuid NOT NULL,
    project_id uuid,
    pid integer,
    name text NOT NULL,
    content text,
    execution_count integer DEFAULT 0,
    run_state text,
    last_run_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    logs text DEFAULT ''::text,
    exception_type text,
    exception_value text
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.scripts OWNER TO postgres;

--
-- TOC entry 398 (class 1259 OID 39016)
-- Name: signals; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.signals (
    id uuid NOT NULL,
    block_id uuid,
    ref_id uuid,
    port_ref_id uuid,
    type text DEFAULT 'INPUT'::text NOT NULL,
    data_type text DEFAULT 'REAL'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text DEFAULT ''::text,
    size integer[],
    is_displayed_on_canvas boolean DEFAULT false,
    is_favorite boolean DEFAULT false,
    is_visible boolean DEFAULT false,
    is_model_port boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    library_data jsonb DEFAULT '{}'::jsonb,
    default_value double precision[],
    CONSTRAINT datatype_not_empty CHECK ((data_type <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.signals OWNER TO postgres;

--
-- TOC entry 399 (class 1259 OID 39033)
-- Name: simulation_configurations; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.simulation_configurations (
    id uuid NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    name text NOT NULL
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.simulation_configurations OWNER TO postgres;

--
-- TOC entry 400 (class 1259 OID 39040)
-- Name: simulation_nodes; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.simulation_nodes (
    id uuid NOT NULL,
    hub_id uuid,
    type text DEFAULT 'VIRTUAL'::text NOT NULL,
    status text DEFAULT 'FREE'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now(),
    version integer DEFAULT 1 NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    description text DEFAULT ''::text,
    ram integer,
    motherboard text,
    cpu text,
    ip inet,
    port integer,
    os text,
    power_controller_ip inet,
    power_controller_port text,
    vm_host_ip inet,
    vm_host_port text,
    dolphin_adapters jsonb,
    io_devices jsonb,
    cores integer DEFAULT 1,
    maintenance_status text DEFAULT ''::text NOT NULL,
    slot_usage jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT status_not_empty CHECK ((status <> ''::text)),
    CONSTRAINT type_not_empty CHECK ((type <> ''::text))
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.simulation_nodes OWNER TO postgres;

--
-- TOC entry 401 (class 1259 OID 39052)
-- Name: simulations; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.simulations (
    id uuid NOT NULL,
    simulation_configuration_id uuid NOT NULL,
    started_at timestamp with time zone DEFAULT now(),
    ended_at timestamp with time zone,
    status text DEFAULT 'CREATED'::text,
    modified_at timestamp with time zone DEFAULT now(),
    target_simulation_node_id uuid,
    username text DEFAULT ''::text NOT NULL,
    hostname text DEFAULT ''::text NOT NULL,
    timestep double precision DEFAULT (0)::double precision NOT NULL,
    stop_sim_timeout double precision DEFAULT 60.0 NOT NULL
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.simulations OWNER TO postgres;

--
-- TOC entry 402 (class 1259 OID 39056)
-- Name: table_datapoints; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.table_datapoints (
    id uuid NOT NULL,
    board_id uuid,
    parameter_id uuid,
    signal_id uuid,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.table_datapoints OWNER TO postgres;

--
-- TOC entry 410 (class 1259 OID 40024)
-- Name: variables; Type: TABLE; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TABLE workspace_s1v8h2yrq9x15u1x.variables (
    id uuid NOT NULL,
    project_id uuid,
    name text NOT NULL,
    value double precision DEFAULT 0.0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    modified_at timestamp with time zone DEFAULT now()
);


ALTER TABLE workspace_s1v8h2yrq9x15u1x.variables OWNER TO postgres;

--
-- TOC entry 4911 (class 2604 OID 38831)
-- Name: migrations id; Type: DEFAULT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.migrations ALTER COLUMN id SET DEFAULT nextval('workspace_s1v8h2yrq9x15u1x.migrations_id_seq'::regclass);


--
-- TOC entry 5568 (class 0 OID 39973)
-- Dependencies: 409
-- Data for Name: aliases; Type: TABLE DATA; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--



--
-- TOC entry 5579 (class 0 OID 0)
-- Dependencies: 381
-- Name: migrations_id_seq; Type: SEQUENCE SET; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

SELECT pg_catalog.setval('workspace_s1v8h2yrq9x15u1x.migrations_id_seq', 173, true);


--
-- TOC entry 5099 (class 2606 OID 39095)
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (id);


--
-- TOC entry 412 (class 1259 OID 40093)
-- Name: interface_block_definition_tree; Type: MATERIALIZED VIEW; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE MATERIALIZED VIEW workspace_s1v8h2yrq9x15u1x.interface_block_definition_tree AS
 WITH lwp AS (
         SELECT f.id,
            f.path,
            f.model_id
           FROM workspace_s1v8h2yrq9x15u1x.get_lwp_all_libraries() f(id uuid, path text, model_id uuid)
        )
 SELECT lwp.model_id,
    b.id,
    b.name,
    b.time_step,
    (b.library_data ->> 'description'::text) AS description,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', p_1.id, 'name', p_1.name, 'type', p_1.default_type, 'direction', p_1.direction, 'dataType', p_1.default_data_type, 'size', p_1.default_size, 'alignment', p_1.alignment, 'linkDirection', p_1.link_direction, 'order', ((p_1.library_data -> 'order'::text))::integer, 'cigreDataType', (p_1.library_data ->> 'cigreDataType'::text))) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.ports p_1
          WHERE (p_1.block_id = b.id)), '[]'::jsonb) AS ports,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', pm.id, 'refId', pm.ref_id, 'name', pm.name, 'isTunable', pm.is_tunable, 'size', pm.size, 'defaultValue', pm.default_value, 'defaultValueText', pm.default_value_text, 'defaultValueObject', pm.default_value_object, 'isDisplayedOnCanvas', pm.is_displayed_on_canvas, 'isFavorite', pm.is_favorite, 'isVisible', pm.is_visible, 'isEqualMode', pm.is_equal_mode, 'isReadOnly', pm.is_read_only, 'dataType', pm.data_type, 'fileId', pm.file_id, 'description', (pm.library_data ->> 'description'::text), 'cigreDataType', (pm.library_data ->> 'cigreDataType'::text), 'isOriginalParameter', (pm.library_data -> 'isOriginalParameter'::text), 'order', ((pm.library_data -> 'order'::text))::integer)) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.parameters pm
          WHERE (pm.block_id = b.id)), '[]'::jsonb) AS parameters,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', s.id, 'refId', s.ref_id, 'name', s.name, 'size', s.size, 'defaultValue', s.default_value, 'type', s.type, 'dataType', s.data_type, 'isDisplayedOnCanvas', s.is_displayed_on_canvas, 'isFavorite', s.is_favorite, 'isVisible', s.is_visible, 'isModelPort', s.is_model_port, 'portRefId', s.port_ref_id, 'description', (s.library_data ->> 'description'::text))) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.signals s
          WHERE (s.block_id = b.id)), '[]'::jsonb) AS signals,
    b.level_id AS parent_id,
    'BLOCK'::text AS type,
        CASE
            WHEN (lwp.path = ''::text) THEN b.name
            ELSE concat(lwp.path, '/', b.name)
        END AS path
   FROM ((((workspace_s1v8h2yrq9x15u1x.blocks b
     LEFT JOIN lwp ON ((lwp.id = b.level_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.models m ON ((m.id = lwp.model_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.projects p ON ((p.id = m.project_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.blocks lb ON ((lb.subsystem_level_id = b.level_id)))
  WHERE ((m.type = 'LIBRARY'::text) AND (p.name = 'Interface Block Libraries'::text) AND ((b.subsystem_level_id IS NULL) OR (b.subsystem_level_id = '00000000-0000-0000-0000-000000000000'::uuid)))
  GROUP BY lwp.model_id, b.id, lwp.path, lb.id
UNION
 SELECT lvl.model_id,
    lvl.id,
    lvl.name,
    NULL::double precision AS time_step,
    NULL::text AS description,
    '[]'::jsonb AS ports,
    '[]'::jsonb AS parameters,
    '[]'::jsonb AS signals,
    lvl.parent_id,
    'LEVEL'::text AS type,
    lwp.path
   FROM ((((workspace_s1v8h2yrq9x15u1x.levels lvl
     LEFT JOIN lwp ON ((lvl.id = lwp.id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.models ON ((models.id = lwp.model_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.projects ON ((projects.id = models.project_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.blocks ON ((blocks.subsystem_level_id = lvl.id)))
  WHERE ((models.type = 'LIBRARY'::text) AND (projects.name = 'Interface Block Libraries'::text))
  ORDER BY 1, 10 DESC
  WITH NO DATA;


ALTER TABLE workspace_s1v8h2yrq9x15u1x.interface_block_definition_tree OWNER TO postgres;

--
-- TOC entry 408 (class 1259 OID 39871)
-- Name: library_block_tree; Type: MATERIALIZED VIEW; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE MATERIALIZED VIEW workspace_s1v8h2yrq9x15u1x.library_block_tree AS
 WITH lwp AS (
         SELECT f.id,
            f.path,
            f.model_id
           FROM workspace_s1v8h2yrq9x15u1x.get_lwp_all_libraries() f(id uuid, path text, model_id uuid)
        )
 SELECT lwp.model_id,
    b.id,
    b.name,
    b.angle,
    b.dimensions,
    b.is_snapped_to_grid,
    (b.library_data ->> 'description'::text) AS description,
    (b.library_data ->> 'geometry'::text) AS geometry,
    (b.library_data ->> 'prefix'::text) AS prefix,
    COALESCE((b.library_data -> 'keywords'::text), '[]'::jsonb) AS keywords,
    (b.library_data -> 'commentOptions'::text) AS comment_options,
    (b.library_data -> 'flipOptions'::text) AS flip_options,
    (b.library_data -> 'resizeOptions'::text) AS resize_options,
    ((b.library_data ->> 'snapToAngle'::text))::double precision AS snap_to_angle,
    ((b.library_data ->> 'isZOrderAllowed'::text))::boolean AS is_z_order_allowed,
    COALESCE((b.library_data -> 'propertyFolders'::text), '[]'::jsonb) AS property_folders,
    COALESCE((b.library_data -> 'callbacks'::text), '[]'::jsonb) AS callbacks,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', p_1.id, 'name', p_1.name, 'type', p_1.default_type, 'direction', p_1.direction, 'dataType', p_1.default_data_type, 'size', p_1.default_size, 'alignment', p_1.alignment, 'linkDirection', p_1.link_direction)) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.ports p_1
          WHERE (p_1.block_id = b.id)), '[]'::jsonb) AS ports,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', pm.id, 'refId', pm.ref_id, 'name', pm.name, 'isTunable', pm.is_tunable, 'size', pm.size, 'defaultValue', pm.default_value, 'defaultValueText', pm.default_value_text, 'isDisplayedOnCanvas', pm.is_displayed_on_canvas, 'isFavorite', pm.is_favorite, 'isVisible', pm.is_visible, 'isEqualMode', pm.is_equal_mode, 'isReadOnly', pm.is_read_only, 'dataType', pm.data_type, 'fileId', pm.file_id, 'libraryData', pm.library_data)) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.parameters pm
          WHERE (pm.block_id = b.id)), '[]'::jsonb) AS parameters,
    COALESCE(( SELECT jsonb_agg(json_build_object('id', s.id, 'refId', s.ref_id, 'name', s.name, 'size', s.size, 'defaultValue', s.default_value, 'type', s.type, 'dataType', s.data_type, 'isDisplayedOnCanvas', s.is_displayed_on_canvas, 'isFavorite', s.is_favorite, 'isVisible', s.is_visible, 'isModelPort', s.is_model_port, 'portRefId', s.port_ref_id)) AS jsonb_agg
           FROM workspace_s1v8h2yrq9x15u1x.signals s
          WHERE (s.block_id = b.id)), '[]'::jsonb) AS signals,
    b.level_id AS parent_id,
    'BLOCK'::text AS type,
        CASE
            WHEN (lwp.path = ''::text) THEN b.name
            ELSE concat(lwp.path, '/', b.name)
        END AS path,
    (b.library_data -> 'icon'::text) AS icon,
    lb.z_order AS level_order,
    b.z_order AS block_order
   FROM ((((workspace_s1v8h2yrq9x15u1x.blocks b
     LEFT JOIN lwp ON ((lwp.id = b.level_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.models m ON ((m.id = lwp.model_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.projects p ON ((p.id = m.project_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.blocks lb ON ((lb.subsystem_level_id = b.level_id)))
  WHERE ((m.type = 'LIBRARY'::text) AND (p.name = 'Native Libraries'::text) AND ((b.subsystem_level_id IS NULL) OR (b.subsystem_level_id = '00000000-0000-0000-0000-000000000000'::uuid)))
  GROUP BY lwp.model_id, b.id, lwp.path, lb.id
UNION
 SELECT lvl.model_id,
    lvl.id,
    lvl.name,
    NULL::double precision AS angle,
    NULL::text AS dimensions,
    NULL::boolean AS is_snapped_to_grid,
    NULL::text AS description,
    NULL::text AS geometry,
    NULL::text AS prefix,
    '[]'::jsonb AS keywords,
    '{}'::jsonb AS comment_options,
    '{}'::jsonb AS flip_options,
    '{}'::jsonb AS resize_options,
    NULL::double precision AS snap_to_angle,
    NULL::boolean AS is_z_order_allowed,
    '[]'::jsonb AS property_folders,
    '[]'::jsonb AS callbacks,
    '[]'::jsonb AS ports,
    '[]'::jsonb AS parameters,
    '[]'::jsonb AS signals,
    lvl.parent_id,
    'LEVEL'::text AS type,
    lwp.path,
    (lvl.metadata -> 'icon'::text) AS icon,
    COALESCE(blocks.z_order, '-1'::integer) AS level_order,
    NULL::integer AS block_order
   FROM ((((workspace_s1v8h2yrq9x15u1x.levels lvl
     LEFT JOIN lwp ON ((lvl.id = lwp.id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.models ON ((models.id = lwp.model_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.projects ON ((projects.id = models.project_id)))
     LEFT JOIN workspace_s1v8h2yrq9x15u1x.blocks ON ((blocks.subsystem_level_id = lvl.id)))
  WHERE ((models.type = 'LIBRARY'::text) AND (projects.name = 'Native Libraries'::text) AND (lvl.id <> '00000000-0000-0000-0000-000000000000'::uuid))
  ORDER BY 1, 25, 22 DESC, 26
  WITH NO DATA;


ALTER TABLE workspace_s1v8h2yrq9x15u1x.library_block_tree OWNER TO postgres;

--
-- TOC entry 5243 (class 2606 OID 40049)
-- Name: aliases aliases_name_unique_project_id; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_name_unique_project_id UNIQUE (project_id, name);


--
-- TOC entry 5246 (class 2606 OID 39984)
-- Name: aliases aliases_parameter_id_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_parameter_id_key UNIQUE (parameter_id);


--
-- TOC entry 5248 (class 2606 OID 39982)
-- Name: aliases aliases_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_pkey PRIMARY KEY (id);


--
-- TOC entry 5252 (class 2606 OID 39986)
-- Name: aliases aliases_signal_id_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_signal_id_key UNIQUE (signal_id);


--
-- TOC entry 5097 (class 2606 OID 39843)
-- Name: blocks blocks_level_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.blocks
    ADD CONSTRAINT blocks_level_id_name_key UNIQUE (level_id, name) DEFERRABLE;


--
-- TOC entry 5141 (class 2606 OID 39099)
-- Name: models board_id_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.models
    ADD CONSTRAINT board_id_unique UNIQUE (board_id);


--
-- TOC entry 5105 (class 2606 OID 39101)
-- Name: boards boards_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.boards
    ADD CONSTRAINT boards_pkey PRIMARY KEY (id);


--
-- TOC entry 5108 (class 2606 OID 39103)
-- Name: boards boards_project_id_parent_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.boards
    ADD CONSTRAINT boards_project_id_parent_id_name_key UNIQUE (project_id, parent_id, name);


--
-- TOC entry 5116 (class 2606 OID 39105)
-- Name: connections connections_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.connections
    ADD CONSTRAINT connections_pkey PRIMARY KEY (id);


--
-- TOC entry 5259 (class 2606 OID 40079)
-- Name: errors errors_parameter_id_parameter_set_id_parameter_raw_index_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.errors
    ADD CONSTRAINT errors_parameter_id_parameter_set_id_parameter_raw_index_unique UNIQUE (parameter_id, parameter_set_id, parameter_raw_index);


--
-- TOC entry 5261 (class 2606 OID 40061)
-- Name: errors errors_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.errors
    ADD CONSTRAINT errors_pkey PRIMARY KEY (id);


--
-- TOC entry 5224 (class 2606 OID 39545)
-- Name: files files_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- TOC entry 5120 (class 2606 OID 39115)
-- Name: hubs hubs_ip_port_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.hubs
    ADD CONSTRAINT hubs_ip_port_unique UNIQUE (ip, port);


--
-- TOC entry 5122 (class 2606 OID 39117)
-- Name: hubs hubs_name_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.hubs
    ADD CONSTRAINT hubs_name_unique UNIQUE (name);


--
-- TOC entry 5124 (class 2606 OID 39119)
-- Name: hubs hubs_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.hubs
    ADD CONSTRAINT hubs_pkey PRIMARY KEY (id);


--
-- TOC entry 5233 (class 2606 OID 39793)
-- Name: io_assignments io_assignments_io_level_id_movel_level_id_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.io_assignments
    ADD CONSTRAINT io_assignments_io_level_id_movel_level_id_key UNIQUE (io_level_id, model_level_id);


--
-- TOC entry 5236 (class 2606 OID 39781)
-- Name: io_assignments io_assignments_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.io_assignments
    ADD CONSTRAINT io_assignments_pkey PRIMARY KEY (id);


--
-- TOC entry 5128 (class 2606 OID 40107)
-- Name: levels levels_model_id_parent_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.levels
    ADD CONSTRAINT levels_model_id_parent_id_name_key UNIQUE (model_id, parent_id, name) DEFERRABLE;


--
-- TOC entry 5132 (class 2606 OID 39123)
-- Name: levels levels_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.levels
    ADD CONSTRAINT levels_pkey PRIMARY KEY (id);


--
-- TOC entry 5136 (class 2606 OID 39702)
-- Name: links links_level_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.links
    ADD CONSTRAINT links_level_id_name_key UNIQUE (level_id, name);


--
-- TOC entry 5138 (class 2606 OID 39125)
-- Name: links links_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.links
    ADD CONSTRAINT links_pkey PRIMARY KEY (id);


--
-- TOC entry 5143 (class 2606 OID 39127)
-- Name: models models_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.models
    ADD CONSTRAINT models_pkey PRIMARY KEY (id);


--
-- TOC entry 5228 (class 2606 OID 39728)
-- Name: nodes nodes_model_id_value_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.nodes
    ADD CONSTRAINT nodes_model_id_value_key UNIQUE (model_id, value);


--
-- TOC entry 5230 (class 2606 OID 39714)
-- Name: nodes nodes_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.nodes
    ADD CONSTRAINT nodes_pkey PRIMARY KEY (id);


--
-- TOC entry 5148 (class 2606 OID 39131)
-- Name: parameter_set_values parameter_set_values_parameter_set_id_parameter_id_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_set_values
    ADD CONSTRAINT parameter_set_values_parameter_set_id_parameter_id_key UNIQUE (parameter_set_id, parameter_id);


--
-- TOC entry 5150 (class 2606 OID 39133)
-- Name: parameter_set_values parameter_set_values_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_set_values
    ADD CONSTRAINT parameter_set_values_pkey PRIMARY KEY (id);


--
-- TOC entry 5152 (class 2606 OID 39135)
-- Name: parameter_sets parameter_sets_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_sets
    ADD CONSTRAINT parameter_sets_pkey PRIMARY KEY (id);


--
-- TOC entry 5156 (class 2606 OID 39137)
-- Name: parameter_sets parameter_sets_project_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_sets
    ADD CONSTRAINT parameter_sets_project_id_name_key UNIQUE (project_id, name);


--
-- TOC entry 5160 (class 2606 OID 39139)
-- Name: parameters parameters_block_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameters
    ADD CONSTRAINT parameters_block_id_name_key UNIQUE (block_id, name);


--
-- TOC entry 5164 (class 2606 OID 39141)
-- Name: parameters parameters_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameters
    ADD CONSTRAINT parameters_pkey PRIMARY KEY (id);


--
-- TOC entry 5170 (class 2606 OID 39143)
-- Name: ports ports_name_block_id_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.ports
    ADD CONSTRAINT ports_name_block_id_key UNIQUE (name, block_id);


--
-- TOC entry 5173 (class 2606 OID 39145)
-- Name: ports ports_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.ports
    ADD CONSTRAINT ports_pkey PRIMARY KEY (id);


--
-- TOC entry 5176 (class 2606 OID 39147)
-- Name: processes processes_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.processes
    ADD CONSTRAINT processes_pkey PRIMARY KEY (id);


--
-- TOC entry 5178 (class 2606 OID 39149)
-- Name: processing_slots processing_slots_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.processing_slots
    ADD CONSTRAINT processing_slots_pkey PRIMARY KEY (id);


--
-- TOC entry 5180 (class 2606 OID 39653)
-- Name: projects projects_checksum_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.projects
    ADD CONSTRAINT projects_checksum_key UNIQUE (checksum);


--
-- TOC entry 5182 (class 2606 OID 39151)
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- TOC entry 5185 (class 2606 OID 39949)
-- Name: reservations reservations_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (simulation_id, interface_block_id, processing_slot_id);


--
-- TOC entry 5187 (class 2606 OID 39155)
-- Name: reservations reservations_processing_slot_id_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_processing_slot_id_unique UNIQUE (processing_slot_id);


--
-- TOC entry 5189 (class 2606 OID 39951)
-- Name: reservations reservations_simulation_id_interface_block_id_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_simulation_id_interface_block_id_unique UNIQUE (simulation_id, interface_block_id);


--
-- TOC entry 5238 (class 2606 OID 40103)
-- Name: scripts script_name_unicity_project_id; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.scripts
    ADD CONSTRAINT script_name_unicity_project_id UNIQUE (project_id, name);


--
-- TOC entry 5240 (class 2606 OID 39817)
-- Name: scripts scripts_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.scripts
    ADD CONSTRAINT scripts_pkey PRIMARY KEY (id);


--
-- TOC entry 5192 (class 2606 OID 39159)
-- Name: signals signals_block_id_name_key; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.signals
    ADD CONSTRAINT signals_block_id_name_key UNIQUE (block_id, name);


--
-- TOC entry 5195 (class 2606 OID 39161)
-- Name: signals signals_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.signals
    ADD CONSTRAINT signals_pkey PRIMARY KEY (id);


--
-- TOC entry 5200 (class 2606 OID 39163)
-- Name: simulation_configurations simulation_configs_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_configurations
    ADD CONSTRAINT simulation_configs_pkey PRIMARY KEY (id);


--
-- TOC entry 5202 (class 2606 OID 39165)
-- Name: simulation_configurations simulation_configs_project_id_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_configurations
    ADD CONSTRAINT simulation_configs_project_id_unique UNIQUE (project_id);


--
-- TOC entry 5204 (class 2606 OID 39167)
-- Name: simulation_configurations simulation_configs_project_id_user_id_name_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_configurations
    ADD CONSTRAINT simulation_configs_project_id_user_id_name_unique UNIQUE (project_id, user_id, name);


--
-- TOC entry 5206 (class 2606 OID 39460)
-- Name: simulation_nodes simulation_nodes_hub_id_ip_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_nodes
    ADD CONSTRAINT simulation_nodes_hub_id_ip_unique UNIQUE (hub_id, ip);


--
-- TOC entry 5208 (class 2606 OID 39171)
-- Name: simulation_nodes simulation_nodes_hub_id_name_unique; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_nodes
    ADD CONSTRAINT simulation_nodes_hub_id_name_unique UNIQUE (hub_id, name);


--
-- TOC entry 5211 (class 2606 OID 39173)
-- Name: simulation_nodes simulation_nodes_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_nodes
    ADD CONSTRAINT simulation_nodes_pkey PRIMARY KEY (id);


--
-- TOC entry 5213 (class 2606 OID 39175)
-- Name: simulations simulations_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulations
    ADD CONSTRAINT simulations_pkey PRIMARY KEY (id);


--
-- TOC entry 5216 (class 2606 OID 39177)
-- Name: table_datapoints table_datapoints_parameter_id; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_parameter_id UNIQUE (board_id, parameter_id);


--
-- TOC entry 5219 (class 2606 OID 39179)
-- Name: table_datapoints table_datapoints_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_pkey PRIMARY KEY (id);


--
-- TOC entry 5221 (class 2606 OID 39181)
-- Name: table_datapoints table_datapoints_signal_id; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_signal_id UNIQUE (board_id, signal_id);


--
-- TOC entry 5254 (class 2606 OID 40035)
-- Name: variables unique_project_name; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.variables
    ADD CONSTRAINT unique_project_name UNIQUE (project_id, name);


--
-- TOC entry 5256 (class 2606 OID 40033)
-- Name: variables variables_pkey; Type: CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.variables
    ADD CONSTRAINT variables_pkey PRIMARY KEY (id);


--
-- TOC entry 5244 (class 1259 OID 40005)
-- Name: aliases_parameter_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX aliases_parameter_id_idx ON workspace_s1v8h2yrq9x15u1x.aliases USING btree (parameter_id);


--
-- TOC entry 5249 (class 1259 OID 40004)
-- Name: aliases_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX aliases_project_id_idx ON workspace_s1v8h2yrq9x15u1x.aliases USING btree (project_id);


--
-- TOC entry 5250 (class 1259 OID 40006)
-- Name: aliases_signal_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX aliases_signal_id_idx ON workspace_s1v8h2yrq9x15u1x.aliases USING btree (signal_id);


--
-- TOC entry 5094 (class 1259 OID 39972)
-- Name: blocks_group_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX blocks_group_id_idx ON workspace_s1v8h2yrq9x15u1x.blocks USING btree (group_id);


--
-- TOC entry 5095 (class 1259 OID 39182)
-- Name: blocks_level_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX blocks_level_id_idx ON workspace_s1v8h2yrq9x15u1x.blocks USING btree (level_id);


--
-- TOC entry 5100 (class 1259 OID 39183)
-- Name: blocks_ref_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX blocks_ref_id_idx ON workspace_s1v8h2yrq9x15u1x.blocks USING btree (ref_id);


--
-- TOC entry 5101 (class 1259 OID 39596)
-- Name: blocks_subsystem_level_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX blocks_subsystem_level_id_idx ON workspace_s1v8h2yrq9x15u1x.blocks USING btree (subsystem_level_id);


--
-- TOC entry 5102 (class 1259 OID 39184)
-- Name: boards_null_parent_id; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX boards_null_parent_id ON workspace_s1v8h2yrq9x15u1x.boards USING btree (project_id, name) WHERE (parent_id IS NULL);


--
-- TOC entry 5103 (class 1259 OID 39185)
-- Name: boards_parent_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX boards_parent_id_idx ON workspace_s1v8h2yrq9x15u1x.boards USING btree (parent_id);


--
-- TOC entry 5106 (class 1259 OID 39186)
-- Name: boards_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX boards_project_id_idx ON workspace_s1v8h2yrq9x15u1x.boards USING btree (project_id);


--
-- TOC entry 5109 (class 1259 OID 39187)
-- Name: connections_from_signal_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX connections_from_signal_id_idx ON workspace_s1v8h2yrq9x15u1x.connections USING btree (from_signal_id);


--
-- TOC entry 5110 (class 1259 OID 39762)
-- Name: connections_is_enabled_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX connections_is_enabled_idx ON workspace_s1v8h2yrq9x15u1x.connections USING btree (to_signal_id) WHERE (is_enabled = true);


--
-- TOC entry 5111 (class 1259 OID 39188)
-- Name: connections_null_from_signal_id; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX connections_null_from_signal_id ON workspace_s1v8h2yrq9x15u1x.connections USING btree (parameter_id, from_signal_id, to_signal_id) WHERE (from_signal_id IS NULL);


--
-- TOC entry 5112 (class 1259 OID 39189)
-- Name: connections_null_parameter_id; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX connections_null_parameter_id ON workspace_s1v8h2yrq9x15u1x.connections USING btree (parameter_id, from_signal_id, to_signal_id) WHERE (parameter_id IS NULL);


--
-- TOC entry 5113 (class 1259 OID 39190)
-- Name: connections_null_to_signal_id; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX connections_null_to_signal_id ON workspace_s1v8h2yrq9x15u1x.connections USING btree (parameter_id, from_signal_id, to_signal_id) WHERE (to_signal_id IS NULL);


--
-- TOC entry 5114 (class 1259 OID 39191)
-- Name: connections_parameter_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX connections_parameter_id_idx ON workspace_s1v8h2yrq9x15u1x.connections USING btree (parameter_id);


--
-- TOC entry 5117 (class 1259 OID 39192)
-- Name: connections_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX connections_project_id_idx ON workspace_s1v8h2yrq9x15u1x.connections USING btree (project_id);


--
-- TOC entry 5118 (class 1259 OID 39193)
-- Name: connections_to_signal_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX connections_to_signal_id_idx ON workspace_s1v8h2yrq9x15u1x.connections USING btree (to_signal_id);


--
-- TOC entry 5262 (class 1259 OID 40077)
-- Name: errors_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX errors_project_id_idx ON workspace_s1v8h2yrq9x15u1x.errors USING btree (project_id);


--
-- TOC entry 5225 (class 1259 OID 39551)
-- Name: files_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX files_project_id_idx ON workspace_s1v8h2yrq9x15u1x.files USING btree (project_id);


--
-- TOC entry 5125 (class 1259 OID 39194)
-- Name: hubs_type_unique_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX hubs_type_unique_idx ON workspace_s1v8h2yrq9x15u1x.hubs USING btree (type) WHERE (type = 'LOCAL'::text);


--
-- TOC entry 5263 (class 1259 OID 40101)
-- Name: interface_block_definition_tree_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX interface_block_definition_tree_id_idx ON workspace_s1v8h2yrq9x15u1x.interface_block_definition_tree USING btree (id);


--
-- TOC entry 5264 (class 1259 OID 40100)
-- Name: interface_block_definition_tree_model_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX interface_block_definition_tree_model_id_idx ON workspace_s1v8h2yrq9x15u1x.interface_block_definition_tree USING btree (model_id);


--
-- TOC entry 5231 (class 1259 OID 39794)
-- Name: io_assignments_io_level_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX io_assignments_io_level_id_idx ON workspace_s1v8h2yrq9x15u1x.io_assignments USING btree (io_level_id);


--
-- TOC entry 5234 (class 1259 OID 39795)
-- Name: io_assignments_level_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX io_assignments_level_id_idx ON workspace_s1v8h2yrq9x15u1x.io_assignments USING btree (model_level_id);


--
-- TOC entry 5126 (class 1259 OID 39195)
-- Name: levels_model_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX levels_model_id_idx ON workspace_s1v8h2yrq9x15u1x.levels USING btree (model_id);


--
-- TOC entry 5129 (class 1259 OID 39196)
-- Name: levels_null_parent_id; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX levels_null_parent_id ON workspace_s1v8h2yrq9x15u1x.levels USING btree (model_id, name) WHERE (parent_id IS NULL);


--
-- TOC entry 5130 (class 1259 OID 39197)
-- Name: levels_parent_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX levels_parent_id_idx ON workspace_s1v8h2yrq9x15u1x.levels USING btree (parent_id);


--
-- TOC entry 5133 (class 1259 OID 40120)
-- Name: links_from_port_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX links_from_port_id_idx ON workspace_s1v8h2yrq9x15u1x.links USING btree (from_port_id);


--
-- TOC entry 5134 (class 1259 OID 39198)
-- Name: links_level_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX links_level_id_idx ON workspace_s1v8h2yrq9x15u1x.links USING btree (level_id);


--
-- TOC entry 5139 (class 1259 OID 40121)
-- Name: links_to_port_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX links_to_port_id_idx ON workspace_s1v8h2yrq9x15u1x.links USING btree (to_port_id);


--
-- TOC entry 5144 (class 1259 OID 39199)
-- Name: models_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX models_project_id_idx ON workspace_s1v8h2yrq9x15u1x.models USING btree (project_id);


--
-- TOC entry 5145 (class 1259 OID 39764)
-- Name: models_project_id_name_unique; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX models_project_id_name_unique ON workspace_s1v8h2yrq9x15u1x.models USING btree (project_id, name, type) WHERE (board_id IS NULL);


--
-- TOC entry 5226 (class 1259 OID 39720)
-- Name: nodes_model_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX nodes_model_id_idx ON workspace_s1v8h2yrq9x15u1x.nodes USING btree (model_id);


--
-- TOC entry 5146 (class 1259 OID 39200)
-- Name: parameter_set_values_parameter_set_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameter_set_values_parameter_set_id_idx ON workspace_s1v8h2yrq9x15u1x.parameter_set_values USING btree (parameter_set_id);


--
-- TOC entry 5153 (class 1259 OID 39201)
-- Name: parameter_sets_project_id_control_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX parameter_sets_project_id_control_id_idx ON workspace_s1v8h2yrq9x15u1x.parameter_sets USING btree (project_id, control_id);


--
-- TOC entry 5154 (class 1259 OID 39202)
-- Name: parameter_sets_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameter_sets_project_id_idx ON workspace_s1v8h2yrq9x15u1x.parameter_sets USING btree (project_id);


--
-- TOC entry 5157 (class 1259 OID 39203)
-- Name: parameter_sets_project_id_type_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameter_sets_project_id_type_idx ON workspace_s1v8h2yrq9x15u1x.parameter_sets USING btree (project_id, type);


--
-- TOC entry 5158 (class 1259 OID 39204)
-- Name: parameters_block_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameters_block_id_idx ON workspace_s1v8h2yrq9x15u1x.parameters USING btree (block_id);


--
-- TOC entry 5161 (class 1259 OID 39478)
-- Name: parameters_datatype_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameters_datatype_idx ON workspace_s1v8h2yrq9x15u1x.parameters USING btree (data_type);


--
-- TOC entry 5162 (class 1259 OID 39561)
-- Name: parameters_file_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameters_file_id_idx ON workspace_s1v8h2yrq9x15u1x.parameters USING btree (file_id);


--
-- TOC entry 5165 (class 1259 OID 39205)
-- Name: parameters_ref_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameters_ref_id_idx ON workspace_s1v8h2yrq9x15u1x.parameters USING btree (ref_id);


--
-- TOC entry 5166 (class 1259 OID 39903)
-- Name: parameters_script_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX parameters_script_id_idx ON workspace_s1v8h2yrq9x15u1x.parameters USING btree (script_id);


--
-- TOC entry 5167 (class 1259 OID 39691)
-- Name: ports_associated_port_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX ports_associated_port_id_idx ON workspace_s1v8h2yrq9x15u1x.ports USING btree (associated_port_id);


--
-- TOC entry 5168 (class 1259 OID 39206)
-- Name: ports_block_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX ports_block_id_idx ON workspace_s1v8h2yrq9x15u1x.ports USING btree (block_id);


--
-- TOC entry 5171 (class 1259 OID 39726)
-- Name: ports_node_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX ports_node_id_idx ON workspace_s1v8h2yrq9x15u1x.ports USING btree (node_id);


--
-- TOC entry 5174 (class 1259 OID 39207)
-- Name: ports_ref_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX ports_ref_id_idx ON workspace_s1v8h2yrq9x15u1x.ports USING btree (ref_id);


--
-- TOC entry 5183 (class 1259 OID 39473)
-- Name: projects_rtlab_hypersim_unique_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX projects_rtlab_hypersim_unique_idx ON workspace_s1v8h2yrq9x15u1x.projects USING btree (path, name) WHERE ((type = ANY (ARRAY['RTLAB'::text, 'HYPERSIM'::text])) AND (path <> ''::text));


--
-- TOC entry 5241 (class 1259 OID 39823)
-- Name: scripts_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX scripts_project_id_idx ON workspace_s1v8h2yrq9x15u1x.scripts USING btree (project_id);


--
-- TOC entry 5190 (class 1259 OID 39208)
-- Name: signals_block_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX signals_block_id_idx ON workspace_s1v8h2yrq9x15u1x.signals USING btree (block_id);


--
-- TOC entry 5193 (class 1259 OID 39479)
-- Name: signals_datatype_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX signals_datatype_idx ON workspace_s1v8h2yrq9x15u1x.signals USING btree (data_type);


--
-- TOC entry 5196 (class 1259 OID 39209)
-- Name: signals_port_ref_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX signals_port_ref_id_idx ON workspace_s1v8h2yrq9x15u1x.signals USING btree (port_ref_id);


--
-- TOC entry 5197 (class 1259 OID 39210)
-- Name: signals_ref_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX signals_ref_id_idx ON workspace_s1v8h2yrq9x15u1x.signals USING btree (ref_id);


--
-- TOC entry 5198 (class 1259 OID 39211)
-- Name: signals_type_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX signals_type_idx ON workspace_s1v8h2yrq9x15u1x.signals USING btree (type);


--
-- TOC entry 5209 (class 1259 OID 39212)
-- Name: simulation_nodes_local_unique_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE UNIQUE INDEX simulation_nodes_local_unique_idx ON workspace_s1v8h2yrq9x15u1x.simulation_nodes USING btree (type) WHERE (type = 'LOCAL'::text);


--
-- TOC entry 5214 (class 1259 OID 39213)
-- Name: table_datapoints_board_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX table_datapoints_board_id_idx ON workspace_s1v8h2yrq9x15u1x.table_datapoints USING btree (board_id);


--
-- TOC entry 5217 (class 1259 OID 39214)
-- Name: table_datapoints_parameter_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX table_datapoints_parameter_id_idx ON workspace_s1v8h2yrq9x15u1x.table_datapoints USING btree (parameter_id);


--
-- TOC entry 5222 (class 1259 OID 39215)
-- Name: table_datapoints_signal_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX table_datapoints_signal_id_idx ON workspace_s1v8h2yrq9x15u1x.table_datapoints USING btree (signal_id);


--
-- TOC entry 5257 (class 1259 OID 40041)
-- Name: variables_project_id_idx; Type: INDEX; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE INDEX variables_project_id_idx ON workspace_s1v8h2yrq9x15u1x.variables USING btree (project_id);


--
-- TOC entry 5337 (class 2620 OID 39682)
-- Name: parameter_set_values notify_parameter_set_values_inserted; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER notify_parameter_set_values_inserted AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.parameter_set_values REFERENCING NEW TABLE AS inserted FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_parameter_set_values('inserted');


--
-- TOC entry 5338 (class 2620 OID 39681)
-- Name: parameter_set_values notify_parameter_set_values_updated; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER notify_parameter_set_values_updated AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.parameter_set_values REFERENCING NEW TABLE AS updated FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_parameter_set_values('updated');


--
-- TOC entry 5353 (class 2620 OID 39661)
-- Name: projects notify_projects_deleted; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER notify_projects_deleted AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.projects REFERENCING OLD TABLE AS deleted FOR EACH ROW WHEN (((old.is_auto_export_enabled = true) AND (old.type = ANY (ARRAY['RTLAB'::text, 'HYPERSIM'::text])) AND (old.path <> ''::text))) EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_deleted();


--
-- TOC entry 5354 (class 2620 OID 39659)
-- Name: projects notify_projects_inserted; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER notify_projects_inserted AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.projects REFERENCING NEW TABLE AS inserted FOR EACH ROW WHEN (((new.is_auto_export_enabled = true) AND (new.type = ANY (ARRAY['RTLAB'::text, 'HYPERSIM'::text])) AND (new.path <> ''::text))) EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_inserted();


--
-- TOC entry 5355 (class 2620 OID 39657)
-- Name: projects notify_projects_updated; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER notify_projects_updated AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.projects REFERENCING NEW TABLE AS updated FOR EACH ROW WHEN (((new.is_auto_export_enabled = true) AND (new.type = ANY (ARRAY['RTLAB'::text, 'HYPERSIM'::text])) AND (new.path <> ''::text) AND (((old.exported_at IS NULL) AND (new.exported_at IS NULL)) OR (old.exported_at = new.exported_at)) AND (((old.checksum IS NULL) AND (new.checksum IS NULL)) OR (old.checksum = new.checksum)))) EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.notify_projects_updated();


--
-- TOC entry 5347 (class 2620 OID 39735)
-- Name: ports on_delete_ports; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER on_delete_ports AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.ports REFERENCING OLD TABLE AS modified_ports FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.delete_unreferenced_nodes();


--
-- TOC entry 5348 (class 2620 OID 39736)
-- Name: ports on_update_ports; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER on_update_ports AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.ports REFERENCING OLD TABLE AS modified_ports FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.delete_unreferenced_nodes();


--
-- TOC entry 5352 (class 2620 OID 39534)
-- Name: processing_slots simulation_nodes_slot_usage; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER simulation_nodes_slot_usage AFTER INSERT OR DELETE OR UPDATE ON workspace_s1v8h2yrq9x15u1x.processing_slots FOR EACH ROW EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_slot_usage();


--
-- TOC entry 5370 (class 2620 OID 40009)
-- Name: aliases table_aliases_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_aliases_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.aliases REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5371 (class 2620 OID 40007)
-- Name: aliases table_aliases_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_aliases_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.aliases REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5372 (class 2620 OID 40008)
-- Name: aliases table_aliases_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_aliases_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.aliases REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5345 (class 2620 OID 39216)
-- Name: parameters table_block_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.parameters REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5349 (class 2620 OID 39218)
-- Name: ports table_block_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.ports REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5356 (class 2620 OID 39217)
-- Name: signals table_block_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.signals REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5350 (class 2620 OID 39221)
-- Name: ports table_block_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.ports REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5346 (class 2620 OID 39222)
-- Name: parameters table_block_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.parameters REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5351 (class 2620 OID 39224)
-- Name: ports table_block_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.ports REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5357 (class 2620 OID 39223)
-- Name: signals table_block_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_block_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.signals REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_block_timestamp();


--
-- TOC entry 5331 (class 2620 OID 39226)
-- Name: models table_board_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.models REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5361 (class 2620 OID 39225)
-- Name: table_datapoints table_board_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.table_datapoints REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5332 (class 2620 OID 39228)
-- Name: models table_board_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5362 (class 2620 OID 39227)
-- Name: table_datapoints table_board_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.table_datapoints REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5333 (class 2620 OID 39230)
-- Name: models table_board_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5363 (class 2620 OID 39229)
-- Name: table_datapoints table_board_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_board_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.table_datapoints REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_board_timestamp();


--
-- TOC entry 5316 (class 2620 OID 39231)
-- Name: blocks table_level_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.blocks REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5328 (class 2620 OID 39232)
-- Name: links table_level_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.links REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5317 (class 2620 OID 39233)
-- Name: blocks table_level_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.blocks REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5329 (class 2620 OID 39234)
-- Name: links table_level_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.links REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5318 (class 2620 OID 39235)
-- Name: blocks table_level_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.blocks REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5330 (class 2620 OID 39236)
-- Name: links table_level_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_level_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.links REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_level_timestamp();


--
-- TOC entry 5325 (class 2620 OID 39237)
-- Name: levels table_model_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_model_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.levels REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_model_timestamp();


--
-- TOC entry 5326 (class 2620 OID 39238)
-- Name: levels table_model_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_model_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.levels REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_model_timestamp();


--
-- TOC entry 5327 (class 2620 OID 39239)
-- Name: levels table_model_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_model_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.levels REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_model_timestamp();


--
-- TOC entry 5339 (class 2620 OID 39649)
-- Name: parameter_set_values table_parameter_and_parameter_set_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_parameter_and_parameter_set_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.parameter_set_values REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_parameter_and_parameter_set_timestamp();


--
-- TOC entry 5340 (class 2620 OID 39650)
-- Name: parameter_set_values table_parameter_and_parameter_set_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_parameter_and_parameter_set_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.parameter_set_values REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_parameter_and_parameter_set_timestamp();


--
-- TOC entry 5341 (class 2620 OID 39651)
-- Name: parameter_set_values table_parameter_and_parameter_set_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_parameter_and_parameter_set_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.parameter_set_values REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_parameter_and_parameter_set_timestamp();


--
-- TOC entry 5319 (class 2620 OID 39244)
-- Name: boards table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.boards REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5322 (class 2620 OID 39246)
-- Name: connections table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.connections REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5376 (class 2620 OID 40082)
-- Name: errors table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.errors REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5364 (class 2620 OID 39554)
-- Name: files table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.files REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5334 (class 2620 OID 39243)
-- Name: models table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.models REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5342 (class 2620 OID 39245)
-- Name: parameter_sets table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.parameter_sets REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5367 (class 2620 OID 39826)
-- Name: scripts table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.scripts REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5358 (class 2620 OID 39745)
-- Name: simulation_configurations table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.simulation_configurations REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5373 (class 2620 OID 40044)
-- Name: variables table_project_timestamp_delete; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_delete AFTER DELETE ON workspace_s1v8h2yrq9x15u1x.variables REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5320 (class 2620 OID 39248)
-- Name: boards table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.boards REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5323 (class 2620 OID 39250)
-- Name: connections table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.connections REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5377 (class 2620 OID 40080)
-- Name: errors table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.errors REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5365 (class 2620 OID 39552)
-- Name: files table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.files REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5335 (class 2620 OID 39247)
-- Name: models table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5343 (class 2620 OID 39249)
-- Name: parameter_sets table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.parameter_sets REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5368 (class 2620 OID 39824)
-- Name: scripts table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.scripts REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5359 (class 2620 OID 39746)
-- Name: simulation_configurations table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.simulation_configurations REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5374 (class 2620 OID 40042)
-- Name: variables table_project_timestamp_insert; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_insert AFTER INSERT ON workspace_s1v8h2yrq9x15u1x.variables REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5321 (class 2620 OID 39252)
-- Name: boards table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.boards REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5324 (class 2620 OID 39254)
-- Name: connections table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.connections REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5378 (class 2620 OID 40081)
-- Name: errors table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.errors REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5366 (class 2620 OID 39553)
-- Name: files table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.files REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5336 (class 2620 OID 39251)
-- Name: models table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5344 (class 2620 OID 39253)
-- Name: parameter_sets table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.parameter_sets REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5369 (class 2620 OID 39825)
-- Name: scripts table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.scripts REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5360 (class 2620 OID 39747)
-- Name: simulation_configurations table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.simulation_configurations REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5375 (class 2620 OID 40043)
-- Name: variables table_project_timestamp_update; Type: TRIGGER; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON workspace_s1v8h2yrq9x15u1x.variables REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION workspace_s1v8h2yrq9x15u1x.update_project_timestamp();


--
-- TOC entry 5309 (class 2606 OID 40110)
-- Name: aliases aliases_parameter_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_parameter_id_fkey FOREIGN KEY (parameter_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameters(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5310 (class 2606 OID 39989)
-- Name: aliases aliases_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5311 (class 2606 OID 40115)
-- Name: aliases aliases_signal_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.aliases
    ADD CONSTRAINT aliases_signal_id_fkey FOREIGN KEY (signal_id) REFERENCES workspace_s1v8h2yrq9x15u1x.signals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5265 (class 2606 OID 39919)
-- Name: blocks blocks_group_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.blocks
    ADD CONSTRAINT blocks_group_id_fkey FOREIGN KEY (group_id) REFERENCES workspace_s1v8h2yrq9x15u1x.blocks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5266 (class 2606 OID 39255)
-- Name: blocks blocks_level_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.blocks
    ADD CONSTRAINT blocks_level_id_fkey FOREIGN KEY (level_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5267 (class 2606 OID 39590)
-- Name: blocks blocks_subsystem_level_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.blocks
    ADD CONSTRAINT blocks_subsystem_level_id_fkey FOREIGN KEY (subsystem_level_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5268 (class 2606 OID 39260)
-- Name: boards boards_parent_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.boards
    ADD CONSTRAINT boards_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES workspace_s1v8h2yrq9x15u1x.boards(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5269 (class 2606 OID 39265)
-- Name: boards boards_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.boards
    ADD CONSTRAINT boards_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5270 (class 2606 OID 39270)
-- Name: connections connections_from_signal_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.connections
    ADD CONSTRAINT connections_from_signal_id_fkey FOREIGN KEY (from_signal_id) REFERENCES workspace_s1v8h2yrq9x15u1x.signals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5271 (class 2606 OID 39275)
-- Name: connections connections_parameter_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.connections
    ADD CONSTRAINT connections_parameter_id_fkey FOREIGN KEY (parameter_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameters(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5272 (class 2606 OID 39280)
-- Name: connections connections_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.connections
    ADD CONSTRAINT connections_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5273 (class 2606 OID 39285)
-- Name: connections connections_to_signal_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.connections
    ADD CONSTRAINT connections_to_signal_id_fkey FOREIGN KEY (to_signal_id) REFERENCES workspace_s1v8h2yrq9x15u1x.signals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5313 (class 2606 OID 40067)
-- Name: errors errors_parameter_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.errors
    ADD CONSTRAINT errors_parameter_id_fkey FOREIGN KEY (parameter_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameters(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5314 (class 2606 OID 40072)
-- Name: errors errors_parameter_set_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.errors
    ADD CONSTRAINT errors_parameter_set_id_fkey FOREIGN KEY (parameter_set_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameter_sets(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5315 (class 2606 OID 40062)
-- Name: errors errors_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.errors
    ADD CONSTRAINT errors_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5304 (class 2606 OID 39546)
-- Name: files files_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.files
    ADD CONSTRAINT files_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5306 (class 2606 OID 39782)
-- Name: io_assignments io_assignments_io_level_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.io_assignments
    ADD CONSTRAINT io_assignments_io_level_id_fkey FOREIGN KEY (io_level_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5307 (class 2606 OID 39787)
-- Name: io_assignments io_assignments_level_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.io_assignments
    ADD CONSTRAINT io_assignments_level_id_fkey FOREIGN KEY (model_level_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5274 (class 2606 OID 39305)
-- Name: levels levels_model_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.levels
    ADD CONSTRAINT levels_model_id_fkey FOREIGN KEY (model_id) REFERENCES workspace_s1v8h2yrq9x15u1x.models(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5275 (class 2606 OID 39310)
-- Name: levels levels_parent_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.levels
    ADD CONSTRAINT levels_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5276 (class 2606 OID 39315)
-- Name: links links_from_ports_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.links
    ADD CONSTRAINT links_from_ports_id_fkey FOREIGN KEY (from_port_id) REFERENCES workspace_s1v8h2yrq9x15u1x.ports(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5277 (class 2606 OID 39320)
-- Name: links links_level_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.links
    ADD CONSTRAINT links_level_id_fkey FOREIGN KEY (level_id) REFERENCES workspace_s1v8h2yrq9x15u1x.levels(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5278 (class 2606 OID 39325)
-- Name: links links_to_ports_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.links
    ADD CONSTRAINT links_to_ports_id_fkey FOREIGN KEY (to_port_id) REFERENCES workspace_s1v8h2yrq9x15u1x.ports(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5279 (class 2606 OID 39330)
-- Name: models models_board_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.models
    ADD CONSTRAINT models_board_id_fkey FOREIGN KEY (board_id) REFERENCES workspace_s1v8h2yrq9x15u1x.boards(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5280 (class 2606 OID 39335)
-- Name: models models_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.models
    ADD CONSTRAINT models_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5305 (class 2606 OID 39729)
-- Name: nodes nodes_model_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.nodes
    ADD CONSTRAINT nodes_model_id_fkey FOREIGN KEY (model_id) REFERENCES workspace_s1v8h2yrq9x15u1x.models(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5281 (class 2606 OID 39340)
-- Name: parameter_set_values parameter_set_values_parameter_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_set_values
    ADD CONSTRAINT parameter_set_values_parameter_id_fkey FOREIGN KEY (parameter_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameters(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5282 (class 2606 OID 39345)
-- Name: parameter_set_values parameter_set_values_parameter_set_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_set_values
    ADD CONSTRAINT parameter_set_values_parameter_set_id_fkey FOREIGN KEY (parameter_set_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameter_sets(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5283 (class 2606 OID 39350)
-- Name: parameter_sets parameter_sets_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameter_sets
    ADD CONSTRAINT parameter_sets_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5284 (class 2606 OID 39355)
-- Name: parameters parameters_block_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameters
    ADD CONSTRAINT parameters_block_id_fkey FOREIGN KEY (block_id) REFERENCES workspace_s1v8h2yrq9x15u1x.blocks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5285 (class 2606 OID 39585)
-- Name: parameters parameters_file_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameters
    ADD CONSTRAINT parameters_file_id_fkey FOREIGN KEY (file_id) REFERENCES workspace_s1v8h2yrq9x15u1x.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5286 (class 2606 OID 39932)
-- Name: parameters parameters_script_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.parameters
    ADD CONSTRAINT parameters_script_id_fkey FOREIGN KEY (script_id) REFERENCES workspace_s1v8h2yrq9x15u1x.scripts(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5287 (class 2606 OID 39686)
-- Name: ports ports_associated_port_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.ports
    ADD CONSTRAINT ports_associated_port_id_fkey FOREIGN KEY (associated_port_id) REFERENCES workspace_s1v8h2yrq9x15u1x.ports(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5288 (class 2606 OID 39360)
-- Name: ports ports_block_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.ports
    ADD CONSTRAINT ports_block_id_fkey FOREIGN KEY (block_id) REFERENCES workspace_s1v8h2yrq9x15u1x.blocks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5289 (class 2606 OID 39721)
-- Name: ports ports_node_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.ports
    ADD CONSTRAINT ports_node_id_fkey FOREIGN KEY (node_id) REFERENCES workspace_s1v8h2yrq9x15u1x.nodes(id) ON UPDATE CASCADE;


--
-- TOC entry 5290 (class 2606 OID 39462)
-- Name: processes processes_processing_slot_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.processes
    ADD CONSTRAINT processes_processing_slot_id_fkey FOREIGN KEY (processing_slot_id) REFERENCES workspace_s1v8h2yrq9x15u1x.processing_slots(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5291 (class 2606 OID 39365)
-- Name: processes processes_simulation_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.processes
    ADD CONSTRAINT processes_simulation_id_fkey FOREIGN KEY (simulation_id) REFERENCES workspace_s1v8h2yrq9x15u1x.simulations(id) ON DELETE CASCADE;


--
-- TOC entry 5292 (class 2606 OID 39370)
-- Name: processing_slots processing_slots_simulation_node_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.processing_slots
    ADD CONSTRAINT processing_slots_simulation_node_id_fkey FOREIGN KEY (simulation_node_id) REFERENCES workspace_s1v8h2yrq9x15u1x.simulation_nodes(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5293 (class 2606 OID 39952)
-- Name: reservations reservations_interface_block_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_interface_block_id_fkey FOREIGN KEY (interface_block_id) REFERENCES workspace_s1v8h2yrq9x15u1x.blocks(id) ON DELETE CASCADE;


--
-- TOC entry 5294 (class 2606 OID 39380)
-- Name: reservations reservations_processing_slot_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_processing_slot_id_fkey FOREIGN KEY (processing_slot_id) REFERENCES workspace_s1v8h2yrq9x15u1x.processing_slots(id) ON DELETE CASCADE;


--
-- TOC entry 5295 (class 2606 OID 39385)
-- Name: reservations reservations_simulation_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.reservations
    ADD CONSTRAINT reservations_simulation_id_fkey FOREIGN KEY (simulation_id) REFERENCES workspace_s1v8h2yrq9x15u1x.simulations(id) ON DELETE CASCADE;


--
-- TOC entry 5308 (class 2606 OID 39818)
-- Name: scripts scripts_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.scripts
    ADD CONSTRAINT scripts_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5296 (class 2606 OID 39390)
-- Name: signals signals_block_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.signals
    ADD CONSTRAINT signals_block_id_fkey FOREIGN KEY (block_id) REFERENCES workspace_s1v8h2yrq9x15u1x.blocks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5297 (class 2606 OID 39480)
-- Name: simulation_configurations simulation_configs_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_configurations
    ADD CONSTRAINT simulation_configs_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5298 (class 2606 OID 39400)
-- Name: simulation_nodes simulation_nodes_hub_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulation_nodes
    ADD CONSTRAINT simulation_nodes_hub_id_fkey FOREIGN KEY (hub_id) REFERENCES workspace_s1v8h2yrq9x15u1x.hubs(id) ON DELETE SET NULL;


--
-- TOC entry 5299 (class 2606 OID 39405)
-- Name: simulations simulations_simulation_config_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulations
    ADD CONSTRAINT simulations_simulation_config_id_fkey FOREIGN KEY (simulation_configuration_id) REFERENCES workspace_s1v8h2yrq9x15u1x.simulation_configurations(id) ON DELETE CASCADE;


--
-- TOC entry 5300 (class 2606 OID 39489)
-- Name: simulations simulations_target_simulation_node_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.simulations
    ADD CONSTRAINT simulations_target_simulation_node_id_fkey FOREIGN KEY (target_simulation_node_id) REFERENCES workspace_s1v8h2yrq9x15u1x.simulation_nodes(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- TOC entry 5301 (class 2606 OID 39410)
-- Name: table_datapoints table_datapoints_board_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_board_id_fkey FOREIGN KEY (board_id) REFERENCES workspace_s1v8h2yrq9x15u1x.boards(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5302 (class 2606 OID 39415)
-- Name: table_datapoints table_datapoints_parameter_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_parameter_id_fkey FOREIGN KEY (parameter_id) REFERENCES workspace_s1v8h2yrq9x15u1x.parameters(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5303 (class 2606 OID 39420)
-- Name: table_datapoints table_datapoints_signal_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.table_datapoints
    ADD CONSTRAINT table_datapoints_signal_id_fkey FOREIGN KEY (signal_id) REFERENCES workspace_s1v8h2yrq9x15u1x.signals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5312 (class 2606 OID 40036)
-- Name: variables variables_project_id_fkey; Type: FK CONSTRAINT; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

ALTER TABLE ONLY workspace_s1v8h2yrq9x15u1x.variables
    ADD CONSTRAINT variables_project_id_fkey FOREIGN KEY (project_id) REFERENCES workspace_s1v8h2yrq9x15u1x.projects(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- TOC entry 5571 (class 0 OID 40093)
-- Dependencies: 412 5573
-- Name: interface_block_definition_tree; Type: MATERIALIZED VIEW DATA; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

REFRESH MATERIALIZED VIEW workspace_s1v8h2yrq9x15u1x.interface_block_definition_tree;


--
-- TOC entry 5567 (class 0 OID 39871)
-- Dependencies: 408 5573
-- Name: library_block_tree; Type: MATERIALIZED VIEW DATA; Schema: workspace_s1v8h2yrq9x15u1x; Owner: postgres
--

REFRESH MATERIALIZED VIEW workspace_s1v8h2yrq9x15u1x.library_block_tree;


-- Completed on 2025-06-27 08:37:52

--
-- PostgreSQL database dump complete
--

