// Lab 진행 단계 목록 컴포넌트.
// steps: Lab DSL에서 가져온 단계 타이틀 / progress: 백엔드 GET /sessions/:id/steps 응답.
// 두 배열을 step_id로 매칭해 각 단계의 상태(pending/active/passed/failed)를 아이콘으로 표시.
import type { StepProgress } from '@/lib/types';

interface Step {
  id: number;
  title: string;
}

interface Props {
  steps: Step[];
  progress: StepProgress[];
  currentStep: number;
}

const STATUS_CONFIG = {
  passed: { icon: '✓', classes: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30' },
  active: { icon: '→', classes: 'text-brand-400 bg-brand-500/10 border-brand-500/30' },
  failed: { icon: '✗', classes: 'text-red-400 bg-red-500/10 border-red-500/30' },
  pending: { icon: '○', classes: 'text-slate-500 bg-slate-800/50 border-slate-700' },
};

export function StepList({ steps, progress, currentStep }: Props) {
  function getStatus(stepId: number) {
    return progress.find((p) => p.step_id === stepId)?.status ?? 'pending';
  }

  return (
    <div className="space-y-2">
      {steps.map((step) => {
        const status = getStatus(step.id);
        const cfg = STATUS_CONFIG[status];
        const isActive = step.id === currentStep;

        return (
          <div
            key={step.id}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg border transition-colors ${cfg.classes} ${
              isActive ? 'ring-1 ring-brand-500/50' : ''
            }`}
          >
            <span className="text-sm font-mono w-4 text-center flex-shrink-0">{cfg.icon}</span>
            <span className="text-sm truncate">{step.title}</span>
          </div>
        );
      })}
    </div>
  );
}
