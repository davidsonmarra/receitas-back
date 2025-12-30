# Guia: Modo de Preparo em Markdown

Este guia explica como utilizar o campo `instructions` (modo de preparo) das receitas, que suporta formatação Markdown.

## 📋 Visão Geral

O campo `instructions` permite que você crie modos de preparo ricos e bem formatados usando Markdown. Isso torna as instruções mais legíveis e organizadas tanto no backend quanto no app React Native.

## ✅ Características

- **Opcional**: Você pode criar receitas sem modo de preparo
- **Tamanho**: Mínimo de 10 caracteres, máximo de 10.000 caracteres
- **Formato**: Markdown básico
- **Armazenamento**: Banco de dados PostgreSQL (tipo TEXT)

## 📝 Markdown Suportado

### Cabeçalhos

```markdown
## Modo de Preparo
### Massa
#### Dica Importante
```

### Listas Numeradas

```markdown
1. Primeiro passo
2. Segundo passo
3. Terceiro passo
```

### Listas Não-Numeradas

```markdown
- Item um
- Item dois
- Item três
```

### Listas Aninhadas

```markdown
1. Preparar a massa:
   - 2 xícaras de farinha
   - 1 xícara de açúcar
   - 3 ovos
2. Misturar tudo
3. Assar
```

### Ênfase

```markdown
**Negrito** - para destacar passos importantes
*Itálico* - para observações ou dicas
***Negrito e Itálico*** - para ênfase máxima
```

### Links

```markdown
Veja mais sobre [técnicas de preparo](https://exemplo.com)
```

## 💡 Exemplos Práticos

### Exemplo 1: Bolo Simples

```json
{
  "title": "Bolo de Chocolate",
  "instructions": "## Modo de Preparo\n\n1. **Pré-aqueça** o forno a 180°C\n2. Em uma tigela, misture:\n   - 2 xícaras de farinha\n   - 1 xícara de açúcar\n   - 1/2 xícara de cacau\n3. Adicione 3 ovos e bata bem\n4. Despeje em forma untada\n5. Asse por 40-45 minutos\n\n*Dica:* Use um palito para verificar se está assado!"
}
```

**Renderizado:**

## Modo de Preparo

1. **Pré-aqueça** o forno a 180°C
2. Em uma tigela, misture:
   - 2 xícaras de farinha
   - 1 xícara de açúcar
   - 1/2 xícara de cacau
3. Adicione 3 ovos e bata bem
4. Despeje em forma untada
5. Asse por 40-45 minutos

*Dica:* Use um palito para verificar se está assado!

---

### Exemplo 2: Receita com Seções

```json
{
  "title": "Lasanha à Bolonhesa",
  "instructions": "## Modo de Preparo\n\n### 1. Molho Bolonhesa\n\n1. Refogue a cebola e o alho\n2. Adicione a carne moída\n3. Acrescente o molho de tomate\n4. Tempere com sal, pimenta e manjericão\n5. Deixe cozinhar por 30 minutos\n\n### 2. Molho Branco\n\n1. Derreta a manteiga\n2. Adicione a farinha e mexa bem\n3. Acrescente o leite aos poucos\n4. Tempere com sal e noz-moscada\n\n### 3. Montagem\n\n1. Unte uma forma\n2. Alterne camadas:\n   - Massa de lasanha\n   - Molho bolonhesa\n   - Molho branco\n   - Queijo ralado\n3. Finalize com queijo\n4. Asse a 180°C por 40 minutos\n\n**Importante:** Deixe descansar 10 minutos antes de servir!"
}
```

---

### Exemplo 3: Receita Rápida

```json
{
  "title": "Omelete Simples",
  "instructions": "1. Bata 3 ovos com sal e pimenta\n2. Aqueça uma frigideira com óleo\n3. Despeje os ovos\n4. Deixe cozinhar por 2-3 minutos\n5. Vire e cozinhe por mais 1 minuto\n\n*Opcional:* Adicione queijo, presunto ou legumes!"
}
```

## 🔌 Integração com React Native

Para renderizar o Markdown no app React Native, utilize a biblioteca `react-native-markdown-display`:

### Instalação

```bash
npm install react-native-markdown-display
# ou
yarn add react-native-markdown-display
```

### Uso Básico

```jsx
import React from 'react';
import { ScrollView, StyleSheet } from 'react-native';
import Markdown from 'react-native-markdown-display';

export default function RecipeInstructions({ recipe }) {
  return (
    <ScrollView style={styles.container}>
      <Markdown style={markdownStyles}>
        {recipe.instructions || '*Sem instruções disponíveis*'}
      </Markdown>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
});

const markdownStyles = {
  heading2: {
    fontSize: 20,
    fontWeight: 'bold',
    marginTop: 16,
    marginBottom: 8,
  },
  heading3: {
    fontSize: 18,
    fontWeight: 'bold',
    marginTop: 12,
    marginBottom: 6,
  },
  body: {
    fontSize: 16,
    lineHeight: 24,
  },
  list_item: {
    marginBottom: 8,
  },
  ordered_list: {
    marginLeft: 16,
  },
  bullet_list: {
    marginLeft: 16,
  },
  em: {
    fontStyle: 'italic',
    color: '#666',
  },
  strong: {
    fontWeight: 'bold',
  },
};
```

### Exemplo Completo

```jsx
import React, { useEffect, useState } from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import Markdown from 'react-native-markdown-display';

export default function RecipeDetail({ route }) {
  const { recipeId } = route.params;
  const [recipe, setRecipe] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`https://api.exemplo.com/recipes/${recipeId}`)
      .then(res => res.json())
      .then(data => {
        setRecipe(data);
        setLoading(false);
      });
  }, [recipeId]);

  if (loading) {
    return <ActivityIndicator />;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>{recipe.title}</Text>
      <Text style={styles.description}>{recipe.description}</Text>
      
      {recipe.instructions && (
        <View style={styles.instructionsContainer}>
          <Markdown style={markdownStyles}>
            {recipe.instructions}
          </Markdown>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  description: {
    fontSize: 16,
    color: '#666',
    marginBottom: 16,
  },
  instructionsContainer: {
    marginTop: 16,
    padding: 16,
    backgroundColor: '#f9f9f9',
    borderRadius: 8,
  },
});

const markdownStyles = {
  // ... estilos do exemplo anterior
};
```

## 🎨 Personalização de Estilos

Você pode personalizar completamente a aparência do Markdown:

```jsx
const customMarkdownStyles = {
  // Cabeçalhos
  heading1: { fontSize: 24, fontWeight: 'bold', color: '#1a1a1a' },
  heading2: { fontSize: 20, fontWeight: 'bold', color: '#2c3e50' },
  heading3: { fontSize: 18, fontWeight: '600', color: '#34495e' },
  
  // Texto
  body: { fontSize: 16, lineHeight: 24, color: '#333' },
  paragraph: { marginBottom: 12 },
  
  // Listas
  ordered_list: { marginLeft: 20 },
  bullet_list: { marginLeft: 20 },
  list_item: { marginBottom: 6, fontSize: 16 },
  
  // Ênfase
  strong: { fontWeight: 'bold', color: '#e74c3c' },
  em: { fontStyle: 'italic', color: '#7f8c8d' },
  
  // Links
  link: { color: '#3498db', textDecorationLine: 'underline' },
};
```

## 🔄 Atualizando Instruções

### Criar Receita com Instruções

```bash
curl -X POST http://localhost:8080/recipes \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bolo de Cenoura",
    "description": "Bolo fofinho com cobertura de chocolate",
    "instructions": "## Modo de Preparo\n\n1. Bata no liquidificador as cenouras com os ovos e o óleo\n2. Despeje em uma tigela\n3. Adicione o açúcar e a farinha\n4. Por último, o fermento\n5. Asse em forma untada a 180°C por 40 minutos",
    "prep_time": 60,
    "servings": 12,
    "difficulty": "fácil"
  }'
```

### Atualizar Instruções Existentes

```bash
curl -X PUT http://localhost:8080/recipes/123 \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "instructions": "## Modo de Preparo Atualizado\n\n1. Novo passo 1\n2. Novo passo 2"
  }'
```

## ✨ Boas Práticas

### ✅ Faça:
- Use listas numeradas para passos sequenciais
- Use **negrito** para destacar ações importantes
- Use *itálico* para dicas e observações
- Organize receitas complexas com cabeçalhos (##, ###)
- Mantenha os passos claros e concisos

### ❌ Evite:
- Textos muito longos em um único passo
- Formatação excessiva que pode dificultar a leitura
- Caracteres especiais desnecessários
- Instruções ambíguas ou vagas

## 📱 Preview Visual

Ao renderizar no app, o usuário verá algo assim:

```
┌─────────────────────────────────┐
│                                 │
│  ## Modo de Preparo             │
│                                 │
│  1. Pré-aqueça o forno          │
│  2. Misture os ingredientes     │
│     • 2 xícaras de farinha      │
│     • 1 xícara de açúcar        │
│  3. Asse por 40 minutos         │
│                                 │
│  Dica: Use forma untada!        │
│                                 │
└─────────────────────────────────┘
```

## 🚀 Próximos Passos

1. Aplicar a migração do banco de dados (ver `migrations/README.md`)
2. Testar criação de receitas com instruções
3. Implementar renderização no React Native
4. (Opcional) Adicionar preview de Markdown no formulário de criação

## 📚 Recursos Adicionais

- [Markdown Guide](https://www.markdownguide.org/)
- [react-native-markdown-display](https://github.com/iamacup/react-native-markdown-display)
- [CommonMark Spec](https://commonmark.org/)

