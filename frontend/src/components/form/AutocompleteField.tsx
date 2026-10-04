import { useFormContext, useWatch, type FieldPath, type FieldValues } from 'react-hook-form';
import {
  Autocomplete,
  AutocompleteContent,
  AutocompleteEmpty,
  AutocompleteInput,
  AutocompleteInputGroup,
  AutocompleteItem,
  AutocompleteList,
  AutocompleteTrigger,
} from '@/components/input';
import { Field, fieldDescribedBy } from './Field';
import type { SelectOption } from './SelectField';

const CREATE_VALUE = '__create_new__';

interface AutocompleteFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  queryName: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  placeholder?: string;
  className?: string;
  options: SelectOption[];
  createOption?: { label: string; onCreate: () => void };
  onSelect?: (value: string) => void;
}

// AutocompleteField binds two form fields to one Base UI
// autocomplete: queryName keeps the text shown in the input (so it stays
// editable while filtering) and name keeps the value of the entry picked
// from the list. A query that matches no option resolves to an empty
// name, which lets the schema require a real selection.
//
// The createOption entry is only offered while the query matches nothing,
// and it acts on an explicit press of the entry — typing never triggers
// it. Base UI commits a highlighted entry with a real click, so pressing
// it with the keyboard counts as a press too.
export function AutocompleteField<T extends FieldValues>({
  name,
  queryName,
  label,
  description,
  required,
  placeholder = 'Escriba para buscar…',
  className,
  options,
  createOption,
  onSelect,
}: AutocompleteFieldProps<T>) {
  const { control, setValue, formState } = useFormContext<T>();
  const query = (useWatch({ control, name: queryName }) as string) ?? '';
  const error = formState.errors[name]?.message as string | undefined;
  const id = String(name);
  const set = (field: FieldPath<T>, value: string) => setValue(field, value as never);

  const normalize = (text: string) =>
    text
      .trim()
      .toLowerCase()
      .normalize('NFD')
      .replace(/\p{Diacritic}/gu, '');
  const match = (text: string) => options.find((opt) => normalize(opt.label) === normalize(text));
  const typed = query.trim();
  const createEntry: SelectOption[] =
    createOption && typed !== '' && !match(query) ? [{ value: CREATE_VALUE, label: `+ ${createOption.label}` }] : [];
  const items: SelectOption[] = [...createEntry, ...options];

  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={id}>
      <Autocomplete
        items={items}
        value={query}
        openOnInputClick
        filter={(item: SelectOption, search: string) =>
          item.value === CREATE_VALUE || normalize(item.label).includes(normalize(search))
        }
        itemToStringValue={(item: SelectOption) => (item.value === CREATE_VALUE ? query : item.label)}
        onValueChange={(text) => {
          const option = match(text);
          if (option) {
            set(name, option.value);
            set(queryName, option.label);
            onSelect?.(option.value);
            return;
          }
          set(queryName, text);
          set(name, '');
        }}
      >
        <AutocompleteInputGroup className={error ? 'input-affix--invalid' : undefined}>
          <AutocompleteInput
            id={id}
            invalid={!!error}
            placeholder={placeholder}
            aria-invalid={!!error || undefined}
            aria-required={required || undefined}
            aria-describedby={fieldDescribedBy(id, description, error)}
          />
          <AutocompleteTrigger aria-label="Mostrar opciones" />
        </AutocompleteInputGroup>
        <AutocompleteContent>
          <AutocompleteList>
            {(item: SelectOption) => (
              <AutocompleteItem
                key={item.value}
                value={item}
                className={item.value === CREATE_VALUE ? 'select-item--create' : undefined}
                onClick={item.value === CREATE_VALUE ? () => createOption?.onCreate() : undefined}
              >
                {item.label}
              </AutocompleteItem>
            )}
          </AutocompleteList>
          <AutocompleteEmpty>Sin coincidencias</AutocompleteEmpty>
        </AutocompleteContent>
      </Autocomplete>
    </Field>
  );
}
