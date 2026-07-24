import React from 'react';

interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  badge?: string;
}

export const FeatureCard: React.FC<FeatureCardProps> = ({ icon, title, description, badge }) => {
  return (
    <div className="group relative rounded-2xl border border-slate-800 bg-[#0d111a] p-6 transition-all duration-300 hover:border-zyra-500/50 hover:bg-[#111624] hover:shadow-xl hover:shadow-zyra-500/5">
      <div className="flex items-center justify-between mb-4">
        <div className="w-10 h-10 rounded-xl bg-zyra-500/10 border border-zyra-500/20 text-zyra-400 flex items-center justify-center group-hover:scale-110 group-hover:bg-zyra-500 group-hover:text-black transition-all duration-300">
          {icon}
        </div>
        {badge && (
          <span className="text-[10px] font-mono font-semibold px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700">
            {badge}
          </span>
        )}
      </div>
      <h3 className="text-lg font-bold text-white mb-2 group-hover:text-zyra-300 transition-colors">
        {title}
      </h3>
      <p className="text-sm text-slate-400 leading-relaxed">
        {description}
      </p>
    </div>
  );
};
