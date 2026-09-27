export default function Field({ id, label, type, value, onChange, autoComplete, hint, invalid }) {
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        name={id}
        type={type}
        value={value}
        onChange={onChange}
        autoComplete={autoComplete}
        required
        aria-invalid={invalid || undefined}
        aria-describedby={hint ? `${id}-hint` : undefined}
      />
      {hint && <span className="field-hint" id={`${id}-hint`}>{hint}</span>}
    </div>
  )
}
