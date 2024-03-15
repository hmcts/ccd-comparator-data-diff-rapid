
-- Create table for event data report
create TABLE IF NOT EXISTS public.event_data_report (
    id SERIAL,
    event_id bigint,
    reference VARCHAR(255),
    event_name VARCHAR(70),
    field_name VARCHAR(255),
    previous_event_id bigint,
    previous_event_created_date timestamp,
    previous_event_user_id varchar(64),
    event_created_date timestamp,
    event_user_id varchar(64),
    event_delta bigint,
    old_record text,
    new_record text,
    analyze_result_detail text,
    change_type VARCHAR(255),
    potential_risk BOOLEAN NOT NULL DEFAULT FALSE,
    case_type_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
alter sequence public.event_data_report_id_seq OWNED BY public.event_data_report.id CACHE 50;
