export const SECURITY_QUESTION_CUSTOM = '__custom__';

export const SECURITY_QUESTIONS: { value: string; label: string }[] = [
  { value: 'mother-maiden', label: '¿Cuál es el apellido de soltera de tu madre?' },
  { value: 'pet', label: '¿Cuál es el nombre de tu primera mascota?' },
  { value: 'birth-city', label: '¿En qué ciudad naciste?' },
  { value: 'school', label: '¿Cuál es el nombre de tu escuela primaria?' },
  { value: 'mother-name', label: '¿Cuál es el nombre de tu madre?' },
  { value: 'best-friend', label: '¿Cómo se llama tu mejor amigo de la infancia?' },
];

export function securityQuestionLabel(value: string): string {
  if (value === SECURITY_QUESTION_CUSTOM) return 'Pregunta personalizada';
  return SECURITY_QUESTIONS.find((q) => q.value === value)?.label ?? value;
}