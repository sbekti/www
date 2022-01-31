FROM nginx:alpine
COPY src /usr/share/nginx/html
RUN find /usr/share/nginx/html/ -type f -exec sed -i -e "s/\(<span class=\"last-updated\">\).*\(<\/span>\)/<span class=\"last-updated\">$(date)<\/span>/g" {} \;