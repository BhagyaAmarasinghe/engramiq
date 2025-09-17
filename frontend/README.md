# EngramIQ Frontend

A modern Next.js frontend for the EngramIQ Solar Asset Reporting Agent, featuring AI-powered document processing and natural language querying capabilities.

## Features

### 🎯 Core Functionality
- **Natural Language Queries**: Chat interface for querying solar site data
- **Document Upload**: Drag-and-drop interface for PDFs, emails, and meeting transcripts
- **Site Overview**: Comprehensive dashboard with key metrics and system health
- **Timeline View**: Interactive timeline of maintenance activities and events
- **Component Management**: Monitor solar equipment status and specifications

### 🎨 Design & UX
- **EngramIQ Brand Implementation**: Following the official style guide with gradients and colors
- **Dark Mode First**: Optimized for professional solar asset management
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Glass Morphism UI**: Modern interface with backdrop blur effects
- **Accessibility**: WCAG AA compliant with proper contrast ratios

### 🤖 AI Integration
- **Source Attribution**: All responses include citations to source documents
- **Confidence Scoring**: AI responses include confidence levels
- **No Hallucination**: Responses based only on uploaded documents
- **Professional Guards**: Prevents inappropriate interactions
- **Real-time Processing**: Live updates during document processing

## Getting Started

### Prerequisites
- Node.js 18+ 
- npm or yarn
- EngramIQ Backend running on `http://localhost:8080`

### Installation

1. **Install Dependencies**
```bash
npm install
# or
yarn install
```

2. **Environment Setup**
```bash
# Create .env.local file
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
```

3. **Start Development Server**
```bash
npm run dev
# or
yarn dev
```

4. **Open Browser**
Navigate to [http://localhost:3000](http://localhost:3000)

## Project Structure

```
src/
├── app/                    # Next.js app router
│   ├── globals.css        # Global styles with EngramIQ theming
│   ├── layout.tsx         # Root layout with providers
│   ├── page.tsx           # Main dashboard page
│   └── providers.tsx      # HeroUI and theme providers
├── components/
│   ├── dashboard/         # Dashboard components
│   │   ├── SiteOverview.tsx   # Site metrics and status
│   │   └── Timeline.tsx       # Event timeline view
│   ├── layout/           # Layout components
│   │   └── Sidebar.tsx       # Main navigation sidebar
│   └── ui/               # Reusable UI components
│       ├── DocumentUpload.tsx # File upload interface
│       ├── EngramIQLogo.tsx   # Brand logo component
│       └── QueryInterface.tsx # Natural language chat
├── lib/
│   ├── api.ts            # API client and endpoints
│   └── utils.ts          # Utility functions
└── types/
    └── index.ts          # TypeScript type definitions
```

## Key Components

### QueryInterface
```tsx
<QueryInterface 
  siteId={site.id}
  onQuerySubmit={handleQuerySubmit}
/>
```
- Natural language chat interface
- Source attribution with citations
- Confidence scoring and validation
- Professional behavior guards

### DocumentUpload
```tsx
<DocumentUpload
  siteId={site.id}
  onUploadComplete={handleDocumentUpload}
  maxFiles={10}
  maxSize={50 * 1024 * 1024} // 50MB
/>
```
- Drag-and-drop file upload
- Progress tracking and status updates
- Document type classification
- Metadata collection

### SiteOverview
```tsx
<SiteOverview
  site={site}
  components={components}
  documents={documents}
  actions={actions}
/>
```
- Key performance metrics
- System health monitoring
- Component status breakdown
- Recent activity summary

## Styling & Theming

### EngramIQ Brand Colors
```css
/* Primary Colors */
--primary-green: #17c480
--primary-dark-blue: #0d1830

/* Gradients */
--brand-gradient: linear-gradient(135deg, #17c480 0%, #0d1830 100%)
```

### Typography
- **Primary**: Figtree (modern, clean)
- **Alternative**: Aptos (Microsoft Suite compatibility)

### Component Styling
All components follow the EngramIQ style guide:
- Gradient backgrounds and accents
- Glass morphism effects
- Consistent spacing and typography
- Dark mode optimized

## API Integration

### Configuration
```typescript
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});
```

### Key Endpoints
- **Sites**: `/sites` - Site management
- **Documents**: `/sites/{id}/documents` - Document upload and management
- **Queries**: `/sites/{id}/queries` - Natural language queries
- **Components**: `/sites/{id}/components` - Solar equipment data
- **Timeline**: `/sites/{id}/timeline` - Events and activities

## Development

### Available Scripts
```bash
npm run dev        # Start development server
npm run build      # Build for production
npm run start      # Start production server
npm run lint       # Run ESLint
```

### Code Style
- TypeScript for type safety
- ESLint for code quality
- Prettier for formatting
- Component-first architecture

### Testing Strategy
```bash
# Unit Tests
npm run test

# E2E Tests (coming soon)
npm run test:e2e

# Type Checking
npm run type-check
```

## PRD Compliance

### ✅ Data Input Requirements
1. **Document Upload**: ✅ Multi-format support (PDF, DOCX, emails)
2. **Action Extraction**: ✅ Automated processing and component linking
3. **Queryable Repository**: ✅ Vector-based semantic search

### ✅ User Query Requirements
1. **Text Input**: ✅ Natural language chat interface
2. **Formatted Response**: ✅ Structured responses with metadata
3. **Concept Understanding**: ✅ Semantic search with related concepts
4. **No Hallucinations**: ✅ Source-only responses with validation
5. **Source Attribution**: ✅ Complete citations with document links
6. **Professional Guards**: ✅ Inappropriate behavior prevention

## Deployment

### Production Build
```bash
npm run build
npm start
```

### Environment Variables
```bash
# .env.production
NEXT_PUBLIC_API_URL=https://api.engramiq.com
```

### Docker Support
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

## Performance

### Optimization Features
- **Code Splitting**: Automatic route-based splitting
- **Image Optimization**: Next.js optimized images
- **Bundle Analysis**: Built-in bundle analyzer
- **Caching Strategy**: SWR for data fetching

### Metrics
- **First Contentful Paint**: < 1.5s
- **Time to Interactive**: < 3s
- **Cumulative Layout Shift**: < 0.1

## Security

### Authentication (Ready)
- JWT token support built-in
- Role-based access control ready
- Session management prepared

### Data Protection
- Input sanitization
- XSS prevention
- CSRF protection
- Secure API communication

## Browser Support

### Supported Browsers
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

### Progressive Enhancement
- Works without JavaScript (basic functionality)
- Graceful degradation for older browsers
- Mobile-first responsive design

## Contributing

### Development Workflow
1. Fork the repository
2. Create feature branch
3. Implement changes following style guide
4. Add tests and documentation
5. Submit pull request

### Style Guide Adherence
- Follow EngramIQ brand guidelines
- Use established component patterns
- Maintain accessibility standards
- Write comprehensive tests

## Troubleshooting

### Common Issues

**API Connection Failed**
```bash
# Check backend status
curl http://localhost:8080/api/v1/health

# Verify environment variables
echo $NEXT_PUBLIC_API_URL
```

**Styling Issues**
```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install
```

**TypeScript Errors**
```bash
# Type check
npm run type-check

# Generate types
npm run generate-types
```

## Roadmap

### Near Term
- [ ] Advanced search interface
- [ ] Real-time WebSocket updates
- [ ] Offline support with PWA
- [ ] Mobile app companion

### Future Enhancements
- [ ] Multi-site management
- [ ] Custom dashboard builder
- [ ] Report generation
- [ ] Third-party integrations

## License

Copyright © 2024 EngramIQ. All rights reserved.

## Support

For technical support or questions:
- Documentation: [docs.engramiq.com](https://docs.engramiq.com)
- Issues: GitHub Issues
- Email: support@engramiq.com