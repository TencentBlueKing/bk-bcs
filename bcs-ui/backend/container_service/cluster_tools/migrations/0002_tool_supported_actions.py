# Generated by Django 3.2.2 on 2022-05-30 11:47

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('cluster_tools', '0001_initial'),
    ]

    operations = [
        migrations.AddField(
            model_name='tool',
            name='supported_actions',
            field=models.CharField(default='install', max_length=128, verbose_name='组件支持的操作'),
        ),
    ]
