import { Box, Stack, Typography } from '@mui/material'

interface Props {
  icon: React.ReactElement
  title: string
  description?: string
  action?: React.ReactNode
}

// Henüz içerik olmayan ekranlar için ortalanmış simge, başlık ve açıklama.
export default function EmptyState({ icon, title, description, action }: Props) {
  return (
    <Stack spacing={1.5} sx={{ alignItems: 'center', py: { xs: 5, md: 8 }, px: 2, textAlign: 'center' }}>
      <Box
        sx={{
          display: 'grid',
          placeItems: 'center',
          width: 72,
          height: 72,
          borderRadius: '50%',
          bgcolor: 'action.hover',
          color: 'primary.main',
          '& svg': { fontSize: 36 },
        }}
      >
        {icon}
      </Box>
      <Typography variant="h6">{title}</Typography>
      {description && (
        <Typography color="text.secondary" sx={{ maxWidth: 420 }}>
          {description}
        </Typography>
      )}
      {action && <Box sx={{ pt: 1 }}>{action}</Box>}
    </Stack>
  )
}
