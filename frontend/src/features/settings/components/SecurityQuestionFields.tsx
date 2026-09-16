import { useWatch } from 'react-hook-form';
import {
  SelectField,
  TextField,
  type SelectOption,
} from '@/components/form';
import {
  SECURITY_QUESTION_CUSTOM,
  SECURITY_QUESTIONS,
  securityQuestionLabel,
} from '@/constants/securityQuestions';

const options: SelectOption[] = [
  ...SECURITY_QUESTIONS,
  { value: SECURITY_QUESTION_CUSTOM, label: 'Escribir mi propia pregunta' },
];

interface SecurityQuestionFieldsProps {
  required?: boolean;
}

export function SecurityQuestionFields({ required }: SecurityQuestionFieldsProps) {
  const question = useWatch({ name: 'securityQuestion' }) as string | undefined;
  const isCustom = question === SECURITY_QUESTION_CUSTOM;
  return (
    <>
      <SelectField
        name="securityQuestion"
        label="Pregunta de seguridad"
        description="La usarás para recuperar el acceso si olvidas tu contraseña."
        placeholder="Seleccione una pregunta…"
        required={required}
        options={options}
        clearable={false}
      />
      {isCustom ? (
        <TextField
          name="customQuestion"
          label="Escribe tu pregunta"
          description="Solo tú conocerás la respuesta."
          required={required}
        />
      ) : (
        question && <p className="field-hint">{securityQuestionLabel(question)}</p>
      )}
      <TextField
        name="securityAnswer"
        label="Tu respuesta"
        required={required}
        autoComplete="off"
      />
    </>
  );
}